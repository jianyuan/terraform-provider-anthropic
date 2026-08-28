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

type ServiceAccountsDataSourceModel struct {
	ServiceAccounts supertypes.SetNestedObjectValueOf[ServiceAccountModel] `tfsdk:"service_accounts"`
}

func (m *ServiceAccountsDataSourceModel) FromAPI(ctx context.Context, serviceAccounts []apiclient.ServiceAccount) (diags diag.Diagnostics) {
	m.ServiceAccounts = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(serviceAccounts, func(sa apiclient.ServiceAccount, _ int) ServiceAccountModel {
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
	var data ServiceAccountsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var serviceAccounts []apiclient.ServiceAccount
	params := &apiclient.ListServiceAccountsParams{
		Limit: new(int64(100)),
	}

	for {
		page := fwdiag.Merge(apiclient.ReadJSON200(d.client.ListServiceAccountsWithResponse(
			ctx,
			params,
			d.WithAuthTokenRequestEditorFn(),
		)))(&resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		serviceAccounts = append(serviceAccounts, page.Data...)

		if v, err := page.NextPage.Get(); err == nil {
			params.Page = new(v)
		} else {
			break
		}
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, serviceAccounts)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
