package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apijson"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

type WorkspaceMembersDataSourceModel struct {
	Id      types.String                                            `tfsdk:"id"`
	Members supertypes.SetNestedObjectValueOf[WorkspaceMemberModel] `tfsdk:"members" apijson:",computed"`
}

func (m *WorkspaceMembersDataSourceModel) FromAPI(ctx context.Context, members []anthropic.WorkspaceMember) (diags diag.Diagnostics) {
	err := apijson.DecodeComputed(ctx, members, m)
	if err != nil {
		diags.AddError("Failed to decode WorkspaceMembers", err.Error())
	}
	return
}

var _ datasource.DataSource = &WorkspaceMembersDataSource{}
var _ datasource.DataSourceWithConfigure = &WorkspaceMembersDataSource{}

func NewWorkspaceMembersDataSource() datasource.DataSource {
	return &WorkspaceMembersDataSource{}
}

type WorkspaceMembersDataSource struct {
	baseDataSource
}

func (d *WorkspaceMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_members"
}

func (d *WorkspaceMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = workspaceMembersSchema().GetDataSource(ctx)
}

func (d *WorkspaceMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	fwdatasource.List(ctx, func(data *WorkspaceMembersDataSourceModel) (fwdatasource.AutoPager[anthropic.WorkspaceMember], diag.Diagnostics) {
		return d.apiKeyClient.Organization.Workspaces.Members.ListAutoPaging(
			ctx,
			data.Id.ValueString(),
			anthropic.OrganizationWorkspaceMemberListParams{},
		), nil
	}, req, resp)
}
