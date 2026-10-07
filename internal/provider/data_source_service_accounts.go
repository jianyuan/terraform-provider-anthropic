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

type ServiceAccountsDataSourceModel struct {
	ServiceAccounts supertypes.SetNestedObjectValueOf[ServiceAccountModel] `tfsdk:"service_accounts"`
}

func (m *ServiceAccountsDataSourceModel) FromAPI(ctx context.Context, serviceAccounts []anthropic.ServiceAccount) (diags diag.Diagnostics) {
	m.ServiceAccounts = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(serviceAccounts, func(sa anthropic.ServiceAccount, _ int) ServiceAccountModel {
		var mm ServiceAccountModel
		diags.Append(mm.FromAPI(ctx, sa)...)
		return mm
	}))
	return
}

var _ datasource.DataSource = &ServiceAccountsDataSource{}
var _ datasource.DataSourceWithConfigure = &ServiceAccountsDataSource{}

func NewServiceAccountsDataSource() datasource.DataSource {
	return &ServiceAccountsDataSource{}
}

type ServiceAccountsDataSource struct {
	baseDataSource
}

func (d *ServiceAccountsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_accounts"
}

func (d *ServiceAccountsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = serviceAccountsSchema().GetDataSource(ctx)
}

func (d *ServiceAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	fwdatasource.List(ctx, func(_ *ServiceAccountsDataSourceModel) fwdatasource.AutoPager[anthropic.ServiceAccount] {
		return d.authTokenClient.Organization.ServiceAccounts.ListAutoPaging(ctx, anthropic.OrganizationServiceAccountListParams{})
	}, req, resp)
}
