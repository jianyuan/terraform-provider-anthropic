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

type UserDataSourceModel struct {
	Id      types.String `tfsdk:"id" apijson:""`
	Email   types.String `tfsdk:"email" apijson:",computed"`
	Name    types.String `tfsdk:"name" apijson:",computed"`
	Role    types.String `tfsdk:"role" apijson:",computed"`
	AddedAt types.String `tfsdk:"added_at" apijson:",computed"`
}

func (m *UserDataSourceModel) FromAPI(ctx context.Context, data anthropic.OrganizationUser) (diags diag.Diagnostics) {
	err := apijson.DecodeComputed(ctx, data, m)
	if err != nil {
		diags.AddError("Failed to decode User", err.Error())
	}
	return
}

var _ datasource.DataSource = &UserDataSource{}
var _ datasource.DataSourceWithConfigure = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	baseDataSource
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get a user in the Organization.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the User.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email of the User.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the User.",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Organization role of the User.",
				Computed:            true,
			},
			"added_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the User joined the Organization.",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	fwdatasource.Read(ctx, func(data *UserDataSourceModel) (*anthropic.OrganizationUser, error) {
		return d.apiKeyClient.Organization.Users.Get(ctx, data.Id.ValueString())
	}, req, resp)
}
