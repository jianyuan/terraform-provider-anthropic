package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apijson"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
)

type OrganizationDataSourceModel struct {
	ID   types.String `tfsdk:"id" apijson:",computed"`
	Name types.String `tfsdk:"name" apijson:",computed"`
}

func (m *OrganizationDataSourceModel) FromAPI(ctx context.Context, data anthropic.OrganizationInfo) (diags diag.Diagnostics) {
	err := apijson.DecodeComputed(ctx, data, m)
	if err != nil {
		diags.AddError("Failed to decode Organization", err.Error())
	}
	return
}

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
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
		MarkdownDescription: "Retrieve information about the organization associated with the authenticated API key.",

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
		return d.apiKeyClient.Organization.Get(ctx)
	}, req, resp)
}
