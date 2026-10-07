package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
)

var _ datasource.DataSource = &WorkspaceMemberDataSource{}
var _ datasource.DataSourceWithConfigure = &WorkspaceMemberDataSource{}

func NewWorkspaceMemberDataSource() datasource.DataSource {
	return &WorkspaceMemberDataSource{}
}

type WorkspaceMemberDataSource struct {
	baseDataSource
}

func (d *WorkspaceMemberDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_member"
}

func (d *WorkspaceMemberDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = workspaceMemberSchema().GetDataSource(ctx)
}

func (d *WorkspaceMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	fwdatasource.Read(ctx, func(data *WorkspaceMemberModel) (*anthropic.WorkspaceMember, error) {
		return d.apiKeyClient.Organization.Workspaces.Members.Get(
			ctx,
			data.UserId.ValueString(),
			anthropic.OrganizationWorkspaceMemberGetParams{
				WorkspaceID: data.WorkspaceId.ValueString(),
			},
		)
	}, req, resp)
}
