package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

var _ resource.Resource = &EnvironmentResource{}
var _ resource.ResourceWithImportState = &EnvironmentResource{}

type EnvironmentResource struct {
	baseResource
}

func (r *EnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Environment resource. An environment is a cloud container template referenced by Managed Agents sessions. " +
			"Environments are immutable: any change to `name` or `config` forces resource replacement. " +
			"On `terraform destroy`, the API hard-delete is attempted first; if sessions still reference the environment, the resource is archived instead.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Environment.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Environment. Must be unique within the workspace.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.SingleNestedAttribute{
				MarkdownDescription: "Container configuration. Immutable after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Environment kind. Currently always `cloud`.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("cloud"),
					},
					"networking": schema.SingleNestedAttribute{
						MarkdownDescription: "Outbound network access policy for the container.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								MarkdownDescription: "Networking mode: `unrestricted` or `limited`.",
								Required:            true,
							},
							"allowed_hosts": schema.ListAttribute{
								MarkdownDescription: "HTTPS-prefixed hosts the container may reach. Only valid when `type = \"limited\"`. Removing the attribute or setting it to `[]` will tighten/clear the list on update — kept Optional (not Computed) on purpose so security-tightening changes take effect.",
								Optional:            true,
								ElementType:         types.StringType,
							},
							"allow_mcp_servers": schema.BoolAttribute{
								MarkdownDescription: "When true, permits outbound access to MCP server endpoints configured on the agent. Only meaningful when `type = \"limited\"`. Defaults to false.",
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
							},
							"allow_package_managers": schema.BoolAttribute{
								MarkdownDescription: "When true, permits outbound access to public package registries. Only meaningful when `type = \"limited\"`. Defaults to false.",
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
							},
						},
					},
					"packages": schema.SingleNestedAttribute{
						MarkdownDescription: "Packages to pre-install in the container, indexed by package manager.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"apt":   schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "System packages installed via apt."},
							"cargo": schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Rust crates installed via cargo."},
							"gem":   schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Ruby gems."},
							"go":    schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Go modules."},
							"npm":   schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Node.js packages."},
							"pip":   schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Python packages."},
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Environment was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Environment was last updated.",
				Computed:            true,
			},
			"archived_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Environment was archived, or null if not archived.",
				Computed:            true,
			},
		},
	}
}

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EnvironmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, diags := environmentConfigToAPI(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateEnvironmentWithResponse(ctx, apiclient.CreateEnvironmentJSONRequestBody{
		Name:   data.Name.ValueString(),
		Config: cfg,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create environment, got empty response body")
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EnvironmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetEnvironmentWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environment, got error: %s", err))
		return
	}
	if httpResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environment, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read environment, got empty response body")
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is a no-op: every attribute is RequiresReplace.
func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EnvironmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EnvironmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := deleteEnvironmentWithFallback(ctx, r.client, data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
	}
}

// deleteEnvironmentWithFallback hard-deletes the environment. The API spec
// returns 409 Conflict when sessions still reference the environment; in that
// case we fall back to archiving so terraform destroy doesn't dangle. Any
// other non-200/non-404 status surfaces as an error rather than silently
// archiving — that protects users from a transient 5xx accidentally turning
// into a permanent archive.
func deleteEnvironmentWithFallback(ctx context.Context, client environmentDeleter, id string) error {
	httpResp, err := client.DeleteEnvironmentWithResponse(ctx, id)
	if err != nil {
		return fmt.Errorf("Unable to delete environment, got error: %s", err)
	}
	switch httpResp.StatusCode() {
	case http.StatusOK, http.StatusNotFound:
		return nil
	case http.StatusConflict:
		// sessions still reference the environment — fall through to archive.
	default:
		return fmt.Errorf("Unable to delete environment, got status code %d: %s",
			httpResp.StatusCode(), string(httpResp.Body))
	}

	archiveResp, err := client.ArchiveEnvironmentWithResponse(ctx, id)
	if err != nil {
		return fmt.Errorf("Unable to archive environment after delete returned 409, got error: %s", err)
	}
	if archiveResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("Unable to archive environment after delete returned 409, got status code %d: %s",
			archiveResp.StatusCode(), string(archiveResp.Body))
	}
	return nil
}

// environmentDeleter is the subset of *apiclient.ClientWithResponses that
// deleteEnvironmentWithFallback uses. Splitting it out lets the unit test
// substitute a fake without spinning up a full provider.
type environmentDeleter interface {
	DeleteEnvironmentWithResponse(ctx context.Context, environmentId string, reqEditors ...apiclient.RequestEditorFn) (*apiclient.DeleteEnvironmentResponse, error)
	ArchiveEnvironmentWithResponse(ctx context.Context, environmentId string, reqEditors ...apiclient.RequestEditorFn) (*apiclient.ArchiveEnvironmentResponse, error)
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
