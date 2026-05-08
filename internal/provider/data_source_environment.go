package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEnvironmentDataSource() datasource.DataSource {
	return &EnvironmentDataSource{}
}

var _ datasource.DataSource = &EnvironmentDataSource{}

type EnvironmentDataSource struct {
	baseDataSource
}

func (d *EnvironmentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *EnvironmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about an Environment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Environment.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Environment.",
				Computed:            true,
			},
			"config": schema.SingleNestedAttribute{
				MarkdownDescription: "Container configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{Computed: true},
					"networking": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"type":                   schema.StringAttribute{Computed: true},
							"allowed_hosts":          schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"allow_mcp_servers":      schema.BoolAttribute{Computed: true},
							"allow_package_managers": schema.BoolAttribute{Computed: true},
						},
					},
					"packages": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"apt":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"cargo": schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"gem":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"go":    schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"npm":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"pip":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
						},
					},
				},
			},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
			"archived_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *EnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetEnvironmentWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environment, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environment, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read environment, got empty response body")
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
