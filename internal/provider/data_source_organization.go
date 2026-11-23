package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
)

type OrganizationDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (m *OrganizationDataSourceModel) FromAPI(ctx context.Context, org anthropic.OrganizationInfo) (diags diag.Diagnostics) {
	m.ID = fwtypes.StringValue(org.ID, org.JSON.ID)
	m.Name = fwtypes.StringValue(org.Name, org.JSON.Name)
	return
}

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{baseDataSource: baseDataSource{base{requiresApiKey: true}}}
}

var _ datasource.DataSource = &OrganizationDataSource{}
var _ datasource.DataSourceWithConfigure = &OrganizationDataSource{}

type OrganizationDataSource struct {
	baseDataSource
}

func (d *OrganizationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *OrganizationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve information about the organization associated with the authenticated API key." + requiresApiKeyNote,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Organization.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Organization.",
				Computed:            true,
			},
		},
	}
}

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	fwdatasource.Read(ctx, func(_ *OrganizationDataSourceModel) (*anthropic.OrganizationInfo, error) {
		return d.client.Organization.Get(ctx)
	}, req, resp)
}
