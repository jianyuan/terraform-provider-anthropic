package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type OrganizationInvitesDataSourceModel struct {
	Invites supertypes.SetNestedObjectValueOf[OrganizationInviteModel] `tfsdk:"invites"`
}

func (m *OrganizationInvitesDataSourceModel) FromAPI(ctx context.Context, invites []apiclient.Invite) (diags diag.Diagnostics) {
	m.Invites = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(invites, func(invite apiclient.Invite, _ int) OrganizationInviteModel {
		var mm OrganizationInviteModel
		diags.Append(mm.FromAPI(ctx, invite)...)
		return mm
	}))
	return
}

var _ datasource.DataSource = &OrganizationInvitesDataSource{}
var _ datasource.DataSourceWithConfigure = &OrganizationInvitesDataSource{}

func NewOrganizationInvitesDataSource() datasource.DataSource {
	return &OrganizationInvitesDataSource{}
}

type OrganizationInvitesDataSource struct {
	baseDataSource
}

func (d *OrganizationInvitesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_invites"
}

func (d *OrganizationInvitesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = organizationInvitesSchema().GetDataSource(ctx)
}

func (d *OrganizationInvitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationInvitesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var invites []apiclient.Invite
	params := &apiclient.ListInvitesParams{
		Limit: new(int64(100)),
	}

	for {
		page := fwdiag.Merge(apiclient.ReadJSON200(d.client.ListInvitesWithResponse(
			ctx,
			params,
			d.WithApiKeyRequestEditorFn(),
		)))(&resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		invites = append(invites, page.Data...)

		if v, err := page.LastId.Get(); err == nil {
			params.AfterId = new(v)
		} else {
			break
		}
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, invites)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
