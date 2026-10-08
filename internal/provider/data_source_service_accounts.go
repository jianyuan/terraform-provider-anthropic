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

type ServiceAccountsDataSourceModel struct {
	ServiceAccounts supertypes.SetNestedObjectValueOf[ServiceAccountModel] `tfsdk:"service_accounts" apijson:",computed"`
}

func (m *ServiceAccountsDataSourceModel) FromAPI(ctx context.Context, data []anthropic.ServiceAccount) (diags diag.Diagnostics) {
	err := apijson.DecodeComputed(ctx, data, m)
	if err != nil {
		diags.AddError("Failed to decode ServiceAccounts", err.Error())
	}
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
	fwdatasource.List(ctx, func(_ *ServiceAccountsDataSourceModel) (fwdatasource.AutoPager[anthropic.ServiceAccount], diag.Diagnostics) {
		return d.authTokenClient.Organization.ServiceAccounts.ListAutoPaging(ctx, anthropic.OrganizationServiceAccountListParams{}), nil
	}, req, resp)
}
