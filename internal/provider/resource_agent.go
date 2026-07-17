package provider

import (
	"context"
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

// withManagedAgentsBeta overrides the anthropic-beta header for /v1/agents requests.
// The provider-level editor sets agent-api-2026-03-01, but the agent_toolset_20260401
// tool type and the per-tool `configs` allowlist require managed-agents-2026-04-01.
// (extensions.go already tolerates the version field being a number under this beta.)
func withManagedAgentsBeta(_ context.Context, req *http.Request) error {
	req.Header.Set("anthropic-beta", "managed-agents-2026-04-01")
	return nil
}

func NewAgentResource() resource.Resource {
	return &AgentResource{}
}

var _ resource.Resource = &AgentResource{}
var _ resource.ResourceWithImportState = &AgentResource{}

type AgentResource struct {
	baseResource
}

func (r *AgentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *AgentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Anthropic Managed Agent.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Agent.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of the Agent. Increments on each update.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Agent.",
				Required:            true,
			},
			"model": schema.StringAttribute{
				MarkdownDescription: "Model ID for the Agent (e.g. `claude-sonnet-4-5`, `claude-opus-4-5`).",
				Required:            true,
			},
			"system": schema.StringAttribute{
				MarkdownDescription: "System prompt for the Agent.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Human-readable description of the Agent.",
				Optional:            true,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Arbitrary string key/value metadata attached to the Agent.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
		},

		Blocks: map[string]schema.Block{
			"tools": schema.ListNestedBlock{
				MarkdownDescription: "Tools available to the Agent.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Tool type (e.g. `agent_toolset_20260401`).",
							Required:            true,
						},
					},
					Blocks: map[string]schema.Block{
						"configs": agentToolConfigsBlock(),
						"default_config": schema.SingleNestedBlock{
							MarkdownDescription: "Default configuration applied to this tool when instantiated in a session.",
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									MarkdownDescription: "Whether the tool is enabled by default.",
									Optional:            true,
								},
							},
							Blocks: map[string]schema.Block{
								"permission_policy": schema.SingleNestedBlock{
									MarkdownDescription: "Permission policy applied to the tool by default. Common values: `always_allow`, `always_ask`.",
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											MarkdownDescription: "Policy type (e.g. `always_allow`, `always_ask`).",
											Optional:            true,
										},
									},
								},
							},
						},
					},
				},
			},
			"mcp_servers": schema.ListNestedBlock{
				MarkdownDescription: "MCP servers available to the Agent.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the MCP server.",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the MCP server (e.g. `url`).",
							Required:            true,
						},
						"url": schema.StringAttribute{
							MarkdownDescription: "URL of the MCP server.",
							Required:            true,
						},
					},
					Blocks: map[string]schema.Block{
						"configs": agentToolConfigsBlock(),
						"default_config": schema.SingleNestedBlock{
							MarkdownDescription: "Default configuration applied to the auto-generated `mcp_toolset` entry for this server.",
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									MarkdownDescription: "Whether the tool is enabled by default.",
									Optional:            true,
								},
							},
							Blocks: map[string]schema.Block{
								"permission_policy": schema.SingleNestedBlock{
									MarkdownDescription: "Permission policy applied to the tool by default. Common values: `always_allow`, `always_ask`.",
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											MarkdownDescription: "Policy type (e.g. `always_allow`, `always_ask`).",
											Optional:            true,
										},
									},
								},
							},
						},
					},
				},
			},
			"skills": schema.ListNestedBlock{
				MarkdownDescription: "Skills assigned to the Agent.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"skill_id": schema.StringAttribute{
							MarkdownDescription: "ID of the skill.",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Skill type (`anthropic` or `custom`).",
							Required:            true,
						},
						"version": schema.StringAttribute{
							MarkdownDescription: "Version of the skill.",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

// agentToolConfigsBlock is the per-tool override list shared by the `tools` and
// `mcp_servers` blocks. Combined with `default_config.enabled = false` it expresses a
// tool allowlist: disable everything by default, then enable only the named tools.
func agentToolConfigsBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "Per-tool overrides within this toolset. Set `default_config.enabled = false` and enable only the tools you want to allowlist.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Bare tool name as reported by the server (e.g. `list_commits`).",
					Required:            true,
				},
				"enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether this specific tool is enabled.",
					Optional:            true,
				},
			},
		},
	}
}

func (r *AgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AgentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.CreateAgentJSONRequestBody{
		Name:  data.Name.ValueString(),
		Model: data.Model.ValueString(),
	}

	if !data.System.IsNull() {
		system := data.System.ValueString()
		body.System = &system
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		body.Description = &desc
	}

	if md := mapFromTFMap(ctx, data.Metadata, &resp.Diagnostics); md != nil {
		body.Metadata = md
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if len(data.Tools) > 0 || len(data.McpServers) > 0 {
		tools := buildAgentTools(data.Tools, data.McpServers)
		body.Tools = &tools
	}

	if len(data.McpServers) > 0 {
		servers := make([]apiclient.AgentMcpServer, len(data.McpServers))
		for i, s := range data.McpServers {
			servers[i] = apiclient.AgentMcpServer{
				Name: s.Name.ValueString(),
				Type: s.Type.ValueString(),
				Url:  s.Url.ValueString(),
			}
		}
		body.McpServers = &servers
	}

	if len(data.Skills) > 0 {
		skills := make([]apiclient.AgentSkillRequest, len(data.Skills))
		for i, s := range data.Skills {
			skills[i] = apiclient.AgentSkillRequest{
				SkillId: s.SkillId.ValueString(),
				Type:    s.Type.ValueString(),
				Version: s.Version.ValueString(),
			}
		}
		body.Skills = &skills
	}

	httpResp, err := r.client.CreateAgentWithResponse(ctx, body, withManagedAgentsBeta)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create agent, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create agent, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create agent, got empty response body")
		return
	}

	if err := data.Fill(*httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AgentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetAgentWithResponse(ctx, data.Id.ValueString(), withManagedAgentsBeta)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read agent, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read agent, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read agent, got empty response body")
		return
	}

	if err := data.Fill(*httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AgentModel
	var state AgentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	model := data.Model.ValueString()
	body := apiclient.UpdateAgentJSONRequestBody{
		Version: state.Version.ValueString(),
		Name:    &name,
		Model:   &model,
	}

	if !data.System.IsNull() {
		system := data.System.ValueString()
		body.System = &system
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		body.Description = &desc
	}

	if md := mapFromTFMap(ctx, data.Metadata, &resp.Diagnostics); md != nil {
		body.Metadata = md
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tools := buildAgentTools(data.Tools, data.McpServers)
	body.Tools = &tools

	servers := make([]apiclient.AgentMcpServer, len(data.McpServers))
	for i, s := range data.McpServers {
		servers[i] = apiclient.AgentMcpServer{
			Name: s.Name.ValueString(),
			Type: s.Type.ValueString(),
			Url:  s.Url.ValueString(),
		}
	}
	body.McpServers = &servers

	skills := make([]apiclient.AgentSkillRequest, len(data.Skills))
	for i, s := range data.Skills {
		skills[i] = apiclient.AgentSkillRequest{
			SkillId: s.SkillId.ValueString(),
			Type:    s.Type.ValueString(),
			Version: s.Version.ValueString(),
		}
	}
	body.Skills = &skills

	httpResp, err := r.client.UpdateAgentWithResponse(ctx, state.Id.ValueString(), body, withManagedAgentsBeta)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update agent, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update agent, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update agent, got empty response body")
		return
	}

	if err := data.Fill(*httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete archives the Agent — there is no hard-delete endpoint (confirmed
// live: DELETE /v1/agents/{id} 404s immediately; `ant beta:agents` only
// exposes create/retrieve/update/list/archive).
func (r *AgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AgentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.ArchiveAgentWithResponse(ctx, data.Id.ValueString(), withManagedAgentsBeta)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to archive agent, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to archive agent, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
}

func (r *AgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
