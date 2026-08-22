package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

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
	permissionPolicyAttr := func(optional bool) schema.SingleNestedAttribute {
		// permission_policy is filled in by the API on apply (the docs say
		// the toolset defaults to always_ask, MCP toolsets to always_ask,
		// agent_toolset to always_allow), so we must declare it Computed
		// to avoid a "Provider produced inconsistent result after apply"
		// error. UseStateForUnknown keeps the prior state on plans where
		// the user didn't change anything; otherwise the framework would
		// re-mark the value as "(known after apply)" on every refresh.
		return schema.SingleNestedAttribute{
			MarkdownDescription: "Tool execution permission policy. Optional+Computed: omitting it lets the API default fill in (e.g. `always_ask` for MCP, `always_allow` for the agent toolset). To change a policy back to the default, set it explicitly rather than removing the attribute.",
			Optional:            optional,
			Computed:            optional,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					MarkdownDescription: "Permission policy kind, e.g. `always_ask`, `always_allow`, `never`.",
					Required:            true,
				},
			},
		}
	}

	toolConfigAttr := func() schema.SingleNestedAttribute {
		return schema.SingleNestedAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Tool name within the toolset (e.g. `web_fetch`, `bash`). Required when used inside `configs`.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether the tool is enabled.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"permission_policy": permissionPolicyAttr(true),
			},
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Agent resource. An agent is a versioned configuration (model, system prompt, tools, MCP servers, skills, multiagent roster) that sessions reference. " +
			"Updates produce a new version on the API side. " +
			"On `terraform destroy`, the agent is archived (not hard-deleted): existing sessions continue to run, but no new sessions can reference it.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Agent.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Current version of the Agent. Bumped by the API every time the agent is updated.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable name of the Agent.",
				Required:            true,
			},
			"model": schema.SingleNestedAttribute{
				MarkdownDescription: "Claude model that powers the agent.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "Model id, e.g. `claude-opus-4-7`.",
						Required:            true,
					},
					"speed": schema.StringAttribute{
						MarkdownDescription: "Model speed mode (`standard` or `fast`). Optional+Computed: the API always echoes back a concrete value (typically `standard`), so we mark it Computed to avoid \"Provider produced inconsistent result after apply\". Once set, removing the line keeps the prior value; assign explicitly to change.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"version": schema.StringAttribute{
						MarkdownDescription: "Model version pin. Optional+Computed for the same reason as `speed`.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"system": schema.StringAttribute{
				MarkdownDescription: "System prompt that defines the agent's behavior. Pass `null` to clear an existing system prompt.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Free-form description of the agent.",
				Optional:            true,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Arbitrary key-value pairs for tracking. Removed keys are sent as empty strings to delete them per API merge semantics; you cannot store a literal empty string as a value.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators:          metadataValueValidators(),
			},
			"tools": schema.ListNestedAttribute{
				MarkdownDescription: "Tools available to the agent. Each element is one of three variants identified by `type`:\n" +
					"- `agent_toolset_20260401` — Anthropic-managed agent toolset; uses `default_config` and/or `configs`.\n" +
					"- `mcp_toolset` — exposes tools from a declared MCP server; uses `mcp_server_name`, `default_config`, `configs`.\n" +
					"- `custom` — user-defined tool executed by your client; uses `name`, `description`, `input_schema`, `permission_policy`.",
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Tool variant: `agent_toolset_20260401`, `mcp_toolset`, or `custom`.",
							Required:            true,
						},
						"default_config": toolConfigAttr(),
						"configs": schema.ListNestedAttribute{
							MarkdownDescription: "Per-tool overrides. Used by `agent_toolset_20260401` and `mcp_toolset`.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "Tool name.",
										Required:            true,
									},
									"enabled": schema.BoolAttribute{
										MarkdownDescription: "Whether the tool is enabled.",
										Optional:            true,
										Computed:            true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"permission_policy": permissionPolicyAttr(true),
								},
							},
						},
						"mcp_server_name": schema.StringAttribute{
							MarkdownDescription: "Name of the declared MCP server. Required when `type = \"mcp_toolset\"`.",
							Optional:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Custom tool name. Required when `type = \"custom\"`.",
							Optional:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Custom tool description. Required when `type = \"custom\"`.",
							Optional:            true,
						},
						"input_schema": schema.StringAttribute{
							MarkdownDescription: "Custom tool input JSON Schema, encoded as JSON (e.g. via `jsonencode`). Required when `type = \"custom\"`. Compared semantically; whitespace and key order don't trigger diffs.",
							CustomType:          jsontypes.NormalizedType{},
							Optional:            true,
						},
						"permission_policy": permissionPolicyAttr(true),
					},
				},
			},
			"mcp_servers": schema.ListNestedAttribute{
				MarkdownDescription: "MCP servers connected to the agent. Auth credentials are supplied at session creation via vaults.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "MCP server kind. Currently always `url`.",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Local name used in `mcp_toolset` references.",
							Required:            true,
						},
						"url": schema.StringAttribute{
							MarkdownDescription: "MCP server URL. Required for `type = \"url\"`.",
							Optional:            true,
						},
					},
				},
			},
			"skills": schema.ListNestedAttribute{
				MarkdownDescription: "Skills attached to the agent.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Skill kind: `anthropic` (pre-built) or `custom`.",
							Required:            true,
						},
						"skill_id": schema.StringAttribute{
							MarkdownDescription: "Skill identifier (e.g. `xlsx` for Anthropic, `skill_*` for custom).",
							Required:            true,
						},
						"version": schema.StringAttribute{
							MarkdownDescription: "Custom-skill version pin (e.g. `latest`).",
							Optional:            true,
						},
					},
				},
			},
			"multiagent": schema.SingleNestedAttribute{
				MarkdownDescription: "Coordinator declaration listing the agents this agent can delegate to.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Multiagent kind. Currently always `coordinator`.",
						Required:            true,
					},
					"agents": schema.ListNestedAttribute{
						MarkdownDescription: "Roster of agents the coordinator can delegate to.",
						Required:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "Roster entry kind: `agent` or `self`.",
									Required:            true,
								},
								"id": schema.StringAttribute{
									MarkdownDescription: "Agent ID. Required when `type = \"agent\"`.",
									Optional:            true,
								},
								"version": schema.Int64Attribute{
									MarkdownDescription: "Pinned agent version. Optional when `type = \"agent\"`; defaults to latest.",
									Optional:            true,
								},
							},
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Agent was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Agent was last updated.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"archived_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Agent was archived, or null if not archived.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
		Name: data.Name.ValueString(),
	}

	model, d := agentModelToAPI(ctx, data.Model)
	resp.Diagnostics.Append(d...)
	body.Model = model

	if !data.System.IsNull() && !data.System.IsUnknown() {
		v := data.System.ValueString()
		body.System = &v
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		v := data.Description.ValueString()
		body.Description = &v
	}

	tools, d := agentToolsFromList(ctx, data.Tools)
	resp.Diagnostics.Append(d...)
	body.Tools = tools

	servers, d := mcpServersFromList(ctx, data.McpServers)
	resp.Diagnostics.Append(d...)
	body.McpServers = servers

	skills, d := skillsFromList(ctx, data.Skills)
	resp.Diagnostics.Append(d...)
	body.Skills = skills

	ma, d := multiagentFromObject(ctx, data.Multiagent)
	resp.Diagnostics.Append(d...)
	body.Multiagent = ma

	md, d := vaultMetadataFromMap(ctx, data.Metadata)
	resp.Diagnostics.Append(d...)
	body.Metadata = md

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateAgentWithResponse(ctx, body)
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

	resp.Diagnostics.Append(data.Fill(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
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

	httpResp, err := r.client.GetAgentWithResponse(ctx, data.Id.ValueString())
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

	resp.Diagnostics.Append(data.Fill(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AgentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// We build the body as a map so we can send explicit JSON nulls (which the
	// API uses as "clear this field") for system/description.
	body := map[string]any{
		"version": state.Version.ValueInt64(),
	}

	if !plan.Name.Equal(state.Name) {
		body["name"] = plan.Name.ValueString()
	}

	if !plan.Model.Equal(state.Model) {
		m, d := agentModelToAPI(ctx, plan.Model)
		resp.Diagnostics.Append(d...)
		if !resp.Diagnostics.HasError() {
			body["model"] = m
		}
	}

	if !plan.System.Equal(state.System) {
		if plan.System.IsNull() {
			body["system"] = nil
		} else {
			body["system"] = plan.System.ValueString()
		}
	}

	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			body["description"] = nil
		} else {
			body["description"] = plan.Description.ValueString()
		}
	}

	if !plan.Tools.Equal(state.Tools) {
		tools, d := agentToolsFromList(ctx, plan.Tools)
		resp.Diagnostics.Append(d...)
		if tools == nil {
			body["tools"] = []apiclient.AgentTool{}
		} else {
			body["tools"] = *tools
		}
	}

	if !plan.McpServers.Equal(state.McpServers) {
		servers, d := mcpServersFromList(ctx, plan.McpServers)
		resp.Diagnostics.Append(d...)
		if servers == nil {
			body["mcp_servers"] = []apiclient.MCPServer{}
		} else {
			body["mcp_servers"] = *servers
		}
	}

	if !plan.Skills.Equal(state.Skills) {
		skills, d := skillsFromList(ctx, plan.Skills)
		resp.Diagnostics.Append(d...)
		if skills == nil {
			body["skills"] = []apiclient.Skill{}
		} else {
			body["skills"] = *skills
		}
	}

	if !plan.Multiagent.Equal(state.Multiagent) {
		ma, d := multiagentFromObject(ctx, plan.Multiagent)
		resp.Diagnostics.Append(d...)
		if ma == nil {
			body["multiagent"] = nil
		} else {
			body["multiagent"] = ma
		}
	}

	if !plan.Metadata.Equal(state.Metadata) {
		md, d := metadataDiff(ctx, state.Metadata, plan.Metadata)
		resp.Diagnostics.Append(d...)
		if md != nil {
			body["metadata"] = *md
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	raw, err := json.Marshal(body)
	if err != nil {
		resp.Diagnostics.AddError("Encoding error", fmt.Sprintf("Failed to encode update body: %s", err))
		return
	}

	httpResp, err := r.client.UpdateAgentWithBodyWithResponse(ctx, state.Id.ValueString(), "application/json", bytes.NewReader(raw))
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

	resp.Diagnostics.Append(plan.Fill(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AgentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.ArchiveAgentWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to archive agent, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != http.StatusOK && httpResp.StatusCode() != http.StatusNotFound {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to archive agent, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
}

func (r *AgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
