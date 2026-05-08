package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewAgentDataSource() datasource.DataSource {
	return &AgentDataSource{}
}

var _ datasource.DataSource = &AgentDataSource{}

type AgentDataSource struct {
	baseDataSource
}

func (d *AgentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (d *AgentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about an Agent.",

		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Required: true, MarkdownDescription: "ID of the Agent."},
			"version": schema.Int64Attribute{Computed: true},
			"name":    schema.StringAttribute{Computed: true},
			"model": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"id":      schema.StringAttribute{Computed: true},
					"speed":   schema.StringAttribute{Computed: true},
					"version": schema.StringAttribute{Computed: true},
				},
			},
			"system":      schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"metadata":    schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"tools": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{Computed: true},
						"default_config": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"name":    schema.StringAttribute{Computed: true},
								"enabled": schema.BoolAttribute{Computed: true},
								"permission_policy": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{Computed: true},
									},
								},
							},
						},
						"configs": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name":    schema.StringAttribute{Computed: true},
									"enabled": schema.BoolAttribute{Computed: true},
									"permission_policy": schema.SingleNestedAttribute{
										Computed: true,
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{Computed: true},
										},
									},
								},
							},
						},
						"mcp_server_name": schema.StringAttribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"description":     schema.StringAttribute{Computed: true},
						"input_schema":    schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}},
						"permission_policy": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{Computed: true},
							},
						},
					},
				},
			},
			"mcp_servers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{Computed: true},
						"name": schema.StringAttribute{Computed: true},
						"url":  schema.StringAttribute{Computed: true},
					},
				},
			},
			"skills": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type":     schema.StringAttribute{Computed: true},
						"skill_id": schema.StringAttribute{Computed: true},
						"version":  schema.StringAttribute{Computed: true},
					},
				},
			},
			"multiagent": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{Computed: true},
					"agents": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type":    schema.StringAttribute{Computed: true},
								"id":      schema.StringAttribute{Computed: true},
								"version": schema.Int64Attribute{Computed: true},
							},
						},
					},
				},
			},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
			"archived_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *AgentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AgentModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetAgentWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read agent, got error: %s", err))
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
