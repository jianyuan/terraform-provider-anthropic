package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type WorkspaceMembersDataSourceModel struct {
	Id      types.String                                            `tfsdk:"id"`
	Members supertypes.SetNestedObjectValueOf[WorkspaceMemberModel] `tfsdk:"members"`
}

func (m *WorkspaceMembersDataSourceModel) FromAPI(FromAPI context.Context, members []apiclient.WorkspaceMember) (diags diag.Diagnostics) {
	m.Members = supertypes.NewSetNestedObjectValueOfValueSlice(FromAPI, lo.Map(members, func(member apiclient.WorkspaceMember, _ int) WorkspaceMemberModel {
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
	var data WorkspaceMembersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var members []apiclient.WorkspaceMember
	params := &apiclient.ListWorkspaceMembersParams{
		Limit: new(int64(100)),
	}

	for {
		page := fwdiag.Merge(apiclient.ReadJSON200(d.client.ListWorkspaceMembersWithResponse(
			ctx,
			data.Id.ValueString(),
			params,
			d.WithApiKeyRequestEditorFn(),
		)))(&resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		members = append(members, page.Data...)

		if v, err := page.LastId.Get(); err == nil {
			params.AfterId = new(v)
		} else {
			break
		}
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, members)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
