package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type WorkspacesDataSourceModel struct {
	Workspaces supertypes.SetNestedObjectValueOf[WorkspaceModel] `tfsdk:"workspaces"`
}

func (m *WorkspacesDataSourceModel) FromAPI(ctx context.Context, workspaces []anthropic.Workspace) (diags diag.Diagnostics) {
	m.Workspaces = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(workspaces, func(workspace anthropic.Workspace, _ int) WorkspaceModel {
		var mm WorkspaceModel
		diags.Append(mm.FromAPI(ctx, workspace)...)
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
	resp.Schema = workspacesSchema().GetDataSource(ctx)
}

func (d *WorkspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	fwdatasource.List(ctx, func(_ *WorkspacesDataSourceModel) (fwdatasource.AutoPager[anthropic.Workspace], diag.Diagnostics) {
		return d.apiKeyClient.Organization.Workspaces.ListAutoPaging(ctx, anthropic.OrganizationWorkspaceListParams{}), nil
	}, req, resp)
}
