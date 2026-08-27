package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type OrganizationDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (m *OrganizationDataSourceModel) Fill(org apiclient.Organization) (diags diag.Diagnostics) {
	m.ID = types.StringValue(org.Id.String())
	m.Name = types.StringValue(org.Name)
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
	var data OrganizationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetCurrentOrganizationWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
		return
	} else if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got status code: %d", httpResp.StatusCode()))
		return
	} else if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read organization, got empty response body")
		return
	}

	resp.Diagnostics.Append(data.Fill(*httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
