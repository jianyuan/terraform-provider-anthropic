package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
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
	var data WorkspaceMemberModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member := fwdiag.Merge(apiclient.ReadJSON200(d.client.GetWorkspaceMemberWithResponse(
		ctx,
		data.WorkspaceId.ValueString(),
		data.UserId.ValueString(),
		d.WithApiKeyRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *member)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
