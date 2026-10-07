package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
)

var _ datasource.DataSource = &ServiceAccountDataSource{}
var _ datasource.DataSourceWithConfigure = &ServiceAccountDataSource{}

func NewServiceAccountDataSource() datasource.DataSource {
	return &ServiceAccountDataSource{}
}

type ServiceAccountDataSource struct {
	baseDataSource
}

func (d *ServiceAccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *ServiceAccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = serviceAccountSchema().GetDataSource(ctx)
}

func (d *ServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	fwdatasource.Read(ctx, func(data *ServiceAccountModel) (*anthropic.ServiceAccount, error) {
		return d.authTokenClient.Organization.ServiceAccounts.Get(ctx, data.Id.ValueString())
	}, req, resp)
}
