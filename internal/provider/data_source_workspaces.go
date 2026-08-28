package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type WorkspacesDataSourceModel struct {
	Workspaces supertypes.SetNestedObjectValueOf[WorkspaceModel] `tfsdk:"workspaces"`
}

func (m *WorkspacesDataSourceModel) Fill(ctx context.Context, workspaces []apiclient.Workspace) (diags diag.Diagnostics) {
	m.Workspaces = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(workspaces, func(workspace apiclient.Workspace, _ int) WorkspaceModel {
		var mm WorkspaceModel
		diags.Append(mm.Fill(ctx, workspace)...)
		return mm
	}))
	return
}

var _ datasource.DataSource = &WorkspacesDataSource{}
var _ datasource.DataSourceWithConfigure = &WorkspacesDataSource{}

func NewWorkspacesDataSource() datasource.DataSource {
	return &WorkspacesDataSource{}
}

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
				CustomType:          supertypes.NewSetNestedObjectTypeOf[WorkspaceModel](ctx),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "ID of the Workspace.",
							Computed:            true,
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
		Limit: new(int64(100)),
	}

	for {
		page := fwdiag.Merge(apiclient.ReadJSON200(d.client.ListWorkspacesWithResponse(
			ctx,
			params,
		)))(&resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		workspaces = append(workspaces, page.Data...)

		if !page.HasMore || page.LastId.IsNull() || !page.LastId.IsSpecified() {
			break
		}

		params.AfterId = new(page.LastId.MustGet())
	}

	resp.Diagnostics.Append(data.Fill(ctx, workspaces)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
