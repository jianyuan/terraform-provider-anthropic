package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apijson"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

type OrganizationInvitesDataSourceModel struct {
	Invites supertypes.SetNestedObjectValueOf[OrganizationInviteModel] `tfsdk:"invites" apijson:",computed"`
}

func (m *OrganizationInvitesDataSourceModel) FromAPI(ctx context.Context, invites []anthropic.OrganizationInvite) (diags diag.Diagnostics) {
	err := apijson.DecodeComputed(ctx, invites, m)
	if err != nil {
		diags.AddError("Failed to decode OrganizationInvites", err.Error())
	}
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
	fwdatasource.List(ctx, func(_ *OrganizationInvitesDataSourceModel) (fwdatasource.AutoPager[anthropic.OrganizationInvite], diag.Diagnostics) {
		return d.apiKeyClient.Organization.Invites.ListAutoPaging(ctx, anthropic.OrganizationInviteListParams{}), nil
	}, req, resp)
}
