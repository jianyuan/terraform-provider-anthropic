package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (m *WorkspaceMembersDataSourceModel) Fill(ctx context.Context, members []apiclient.WorkspaceMember) (diags diag.Diagnostics) {
	m.Members = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(members, func(member apiclient.WorkspaceMember, _ int) WorkspaceMemberModel {
		var mm WorkspaceMemberModel
		diags.Append(mm.Fill(ctx, member)...)
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
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all members of the workspace.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Workspace.",
				Required:            true,
			},
			"members": schema.SetNestedAttribute{
				MarkdownDescription: "List of members.",
				Computed:            true,
				CustomType:          supertypes.NewSetNestedObjectTypeOf[WorkspaceMemberModel](ctx),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"workspace_id": schema.StringAttribute{
							MarkdownDescription: "ID of the Workspace to which the member belongs.",
							Computed:            true,
						},
						"user_id": schema.StringAttribute{
							MarkdownDescription: "ID of the user who is a member of the Workspace.",
							Computed:            true,
						},
						"workspace_role": schema.StringAttribute{
							MarkdownDescription: "Role of the new Workspace Member. Must be one of `workspace_user`, `workspace_developer`, or `workspace_admin`.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
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
		)))(&resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		members = append(members, page.Data...)

		if !page.HasMore || page.LastId.IsNull() || !page.LastId.IsSpecified() {
			break
		}

		params.AfterId = new(page.LastId.MustGet())
	}

	resp.Diagnostics.Append(data.Fill(ctx, members)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
