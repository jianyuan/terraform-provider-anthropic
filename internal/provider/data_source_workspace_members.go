package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type WorkspaceMembersDataSourceModel struct {
	Id      types.String                                            `tfsdk:"id"`
	Members supertypes.SetNestedObjectValueOf[WorkspaceMemberModel] `tfsdk:"members"`
}

func (m *WorkspaceMembersDataSourceModel) FromAPI(FromAPI context.Context, members []anthropic.WorkspaceMember) (diags diag.Diagnostics) {
	m.Members = supertypes.NewSetNestedObjectValueOfValueSlice(FromAPI, lo.Map(members, func(member anthropic.WorkspaceMember, _ int) WorkspaceMemberModel {
		var mm WorkspaceMemberModel
		diags.Append(mm.FromAPI(FromAPI, member)...)
		return mm
	}))
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
	fwdatasource.List(ctx, func(data *WorkspaceMembersDataSourceModel) fwdatasource.AutoPager[anthropic.WorkspaceMember] {
		return d.apiKeyClient.Organization.Workspaces.Members.ListAutoPaging(
			ctx,
			data.Id.ValueString(),
			anthropic.OrganizationWorkspaceMemberListParams{},
		)
	}, req, resp)
}
