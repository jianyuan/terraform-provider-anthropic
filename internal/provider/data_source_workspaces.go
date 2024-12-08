package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jianyuan/go-utils/ptr"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type WorkspacesDataSourceModel struct {
	Workspaces []WorkspaceModel `tfsdk:"workspaces"`
}

func (m *WorkspacesDataSourceModel) Fill(workspaces []apiclient.Workspace) error {
	m.Workspaces = make([]WorkspaceModel, len(workspaces))
	for i, u := range workspaces {
		if err := m.Workspaces[i].Fill(u); err != nil {
			return err
		}
	}

	return nil
}

func NewWorkspacesDataSource() datasource.DataSource {
	return &WorkspacesDataSource{}
}

var _ datasource.DataSource = &WorkspacesDataSource{}

type WorkspacesDataSource struct {
	baseDataSource
}

func (d *WorkspacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspaces"
}

func (d *WorkspacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all workspaces in the organization.",

		Attributes: map[string]schema.Attribute{
			"workspaces": schema.SetNestedAttribute{
				MarkdownDescription: "List of workspaces.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "ID of the Workspace.",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the Workspace.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "RFC 3339 datetime string indicating when the Workspace was created.",
							Computed:            true,
						},
						"archived_at": schema.StringAttribute{
							MarkdownDescription: "RFC 3339 datetime string indicating when the Workspace was archived, or null if the Workspace is not archived.",
							Computed:            true,
						},
						"display_color": schema.StringAttribute{
							MarkdownDescription: "Hex color code representing the Workspace in the Anthropic Console.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *WorkspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkspacesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var workspaces []apiclient.Workspace
	params := &apiclient.ListWorkspacesParams{
		Limit: ptr.Ptr(100),
	}

	for {
		httpResp, err := d.client.ListWorkspacesWithResponse(
			ctx,
			params,
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got status code: %d", httpResp.StatusCode()))
			return
		}

		if httpResp.JSON200 == nil {
			resp.Diagnostics.AddError("Client Error", "Unable to read, got empty response body")
			return
		}

		workspaces = append(workspaces, httpResp.JSON200.Data...)

		if !httpResp.JSON200.HasMore || httpResp.JSON200.LastId == nil {
			break
		}

		params.AfterId = httpResp.JSON200.LastId
	}

	if err := data.Fill(workspaces); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
