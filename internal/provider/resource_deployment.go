package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/frank-bee/terraform-provider-anthropic/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// withDeploymentsBeta overrides the anthropic-beta header for /v1/deployments
// requests, same as agents/environments.
func withDeploymentsBeta(_ context.Context, req *http.Request) error {
	req.Header.Set("anthropic-beta", "managed-agents-2026-04-01")
	return nil
}

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}

type DeploymentResource struct {
	baseResource
}

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Anthropic Managed Agent Deployment — a cron-schedulable (or manually-run) " +
			"wrapper around an agent + environment that creates a new session per run. There is no hard delete " +
			"for this resource; `terraform destroy` archives it instead.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Deployment.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable name for the Deployment.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of what the Deployment does.",
				Optional:            true,
			},
			"agent_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Agent to deploy. Must exist and not be archived.",
				Required:            true,
			},
			"agent_version": schema.StringAttribute{
				MarkdownDescription: "Pin the Deployment to a specific Agent version. Omit to float to the " +
					"Agent's latest version as of each apply (the API re-pins it internally on create/update; " +
					"this attribute then just reflects whatever version that resolved to).",
				Optional: true,
				Computed: true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Environment defining the container configuration for sessions " +
					"created from this Deployment.",
				Required: true,
			},
			"vault_ids": schema.ListAttribute{
				MarkdownDescription: "Vault IDs for stored credentials the agent can use during sessions created " +
					"from this Deployment. Maximum 50.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Free-form string metadata attached to the Deployment.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"initial_events": schema.ListAttribute{
				MarkdownDescription: "Events sent to each session immediately after creation. Each element is a " +
					"JSON-encoded event object, e.g. `jsonencode({type = \"user.message\", content = [{type = " +
					"\"text\", text = \"...\"}]})`. At least 1, maximum 50.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"resources": schema.ListAttribute{
				MarkdownDescription: "Resources (e.g. memory stores, files) mounted into each session's " +
					"container. Each element is a JSON-encoded resource object, e.g. " +
					"`jsonencode({type = \"memory_store\", memory_store_id = ..., access = \"read_write\", " +
					"instructions = \"...\"})`. Maximum 500.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Lifecycle status of the Deployment (e.g. `active`, `paused`, `archived`). " +
					"Not settable here — use `ant beta:deployments pause/unpause` out of band.",
				Computed: true,
			},
			"archived_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Deployment was archived, if ever.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Deployment was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Deployment was last updated.",
				Computed:            true,
			},
		},

		Blocks: map[string]schema.Block{
			"schedule": schema.SingleNestedBlock{
				MarkdownDescription: "Cron schedule for automatic runs. Omit entirely for a manual-only " +
					"Deployment (triggered via `ant beta:deployments run` / the API) — removing this block from " +
					"an existing config clears any previously-set schedule.",
				Attributes: map[string]schema.Attribute{
					"expression": schema.StringAttribute{
						MarkdownDescription: "5-field POSIX cron expression. Literal wall-clock matching in " +
							"`timezone`. Not statically Required — a `schedule` block with no `expression` set " +
							"is rejected by the API, not the provider, so an entirely-omitted `schedule` block " +
							"doesn't false-positive against Terraform's required-attribute validation (which " +
							"checks nested attribute paths even when the parent block is absent).",
						Optional: true,
					},
					"timezone": schema.StringAttribute{
						MarkdownDescription: "IANA timezone for `expression`. Defaults to `UTC`.",
						Optional:            true,
						Computed:            true,
					},
					"last_run_at": schema.StringAttribute{
						MarkdownDescription: "RFC 3339 datetime string of the last scheduled run, if any.",
						Computed:            true,
					},
					"upcoming_runs_at": schema.ListAttribute{
						MarkdownDescription: "RFC 3339 datetime strings of the next few scheduled runs.",
						Computed:            true,
						ElementType:         types.StringType,
					},
				},
			},
		},
	}
}

// buildAgentRef returns either a bare agent-ID string (float to latest) or a
// {id, version} object (pin to a specific version) — the API's `agent` field
// accepts both shapes on create/update, matching what `ant` sends.
func buildAgentRef(id, version types.String) interface{} {
	if !version.IsNull() && !version.IsUnknown() && version.ValueString() != "" {
		return map[string]interface{}{"id": id.ValueString(), "version": version.ValueString()}
	}
	return id.ValueString()
}

// buildScheduleInput converts the `schedule` block into the API shape. `type`
// is hardcoded to "cron" — the only value the API currently supports — since
// the API requires it on write even though it's not exposed as a settable
// field by `ant`.
func buildScheduleInput(s *DeploymentScheduleModel) *apiclient.DeploymentScheduleInput {
	if s == nil {
		return nil
	}
	out := &apiclient.DeploymentScheduleInput{
		Type:       "cron",
		Expression: s.Expression.ValueString(),
	}
	if !s.Timezone.IsNull() && !s.Timezone.IsUnknown() && s.Timezone.ValueString() != "" {
		out.Timezone = ptrTo(s.Timezone.ValueString())
	}
	return out
}

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.CreateDeploymentJSONRequestBody{
		Name:          data.Name.ValueString(),
		EnvironmentId: data.EnvironmentId.ValueString(),
		Agent:         buildAgentRef(data.AgentId, data.AgentVersion),
		Schedule:      buildScheduleInput(data.Schedule),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		body.Description = ptrTo(data.Description.ValueString())
	}
	body.Metadata = mapFromTFMap(ctx, data.Metadata, &resp.Diagnostics)
	body.VaultIds = buildStringList(ctx, data.VaultIds, &resp.Diagnostics)
	body.InitialEvents = buildJSONList(ctx, data.InitialEvents, &resp.Diagnostics)
	body.Resources = buildJSONList(ctx, data.Resources, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateDeploymentWithResponse(ctx, body, withDeploymentsBeta)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create deployment, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create deployment, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create deployment, got empty response body")
		return
	}

	if err := data.Fill(ctx, *httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetDeploymentWithResponse(ctx, data.Id.ValueString(), withDeploymentsBeta)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read deployment, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read deployment, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read deployment, got empty response body")
		return
	}

	if err := data.Fill(ctx, *httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// deploymentUpdateBody mirrors apiclient.UpdateDeploymentRequest except for
// Schedule: unlike the generated struct (which omits nil fields via
// `omitempty` — fine for "leave alone", but unable to express "clear this"),
// Schedule here has no `omitempty` so a nil pointer marshals to a literal
// JSON `null`. Terraform owns the full desired state of this resource, so
// removing the `schedule` block from config must actively clear it on the
// server rather than silently leaving a stale schedule in place (confirmed
// live: POST .../{id} with `{"schedule": null}` clears it).
type deploymentUpdateBody struct {
	Name          *string                            `json:"name,omitempty"`
	Description   *string                            `json:"description,omitempty"`
	Agent         interface{}                        `json:"agent,omitempty"`
	EnvironmentId *string                            `json:"environment_id,omitempty"`
	InitialEvents *[]map[string]interface{}          `json:"initial_events,omitempty"`
	Resources     *[]map[string]interface{}          `json:"resources,omitempty"`
	Metadata      *map[string]string                 `json:"metadata,omitempty"`
	VaultIds      *[]string                          `json:"vault_ids,omitempty"`
	Schedule      *apiclient.DeploymentScheduleInput `json:"schedule"`
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeploymentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := deploymentUpdateBody{
		Name:          ptrTo(data.Name.ValueString()),
		EnvironmentId: ptrTo(data.EnvironmentId.ValueString()),
		Agent:         buildAgentRef(data.AgentId, data.AgentVersion),
		Schedule:      buildScheduleInput(data.Schedule),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		body.Description = ptrTo(data.Description.ValueString())
	}
	body.Metadata = mapFromTFMap(ctx, data.Metadata, &resp.Diagnostics)
	body.VaultIds = buildStringList(ctx, data.VaultIds, &resp.Diagnostics)
	body.InitialEvents = buildJSONList(ctx, data.InitialEvents, &resp.Diagnostics)
	body.Resources = buildJSONList(ctx, data.Resources, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	raw, err := json.Marshal(body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to encode deployment update: %s", err))
		return
	}

	httpResp, err := r.client.UpdateDeploymentWithBodyWithResponse(ctx, data.Id.ValueString(), "application/json", bytes.NewReader(raw), withDeploymentsBeta)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update deployment, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update deployment, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update deployment, got empty response body")
		return
	}

	if err := data.Fill(ctx, *httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete archives the Deployment — there is no hard-delete endpoint for this
// resource (confirmed live: only create/read/update/archive/pause/unpause/run).
func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeploymentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.ArchiveDeploymentWithResponse(ctx, data.Id.ValueString(), withDeploymentsBeta)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to archive deployment, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to archive deployment, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
