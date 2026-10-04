package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
)

type UserDataSourceModel struct {
	Id      types.String `tfsdk:"id"`
	Email   types.String `tfsdk:"email"`
	Name    types.String `tfsdk:"name"`
	Role    types.String `tfsdk:"role"`
	AddedAt types.String `tfsdk:"added_at"`
}

func (m *UserDataSourceModel) FromAPI(ctx context.Context, data anthropic.OrganizationUser) (diags diag.Diagnostics) {
	m.Id = fwtypes.StringValue(data.ID, data.JSON.ID)
	m.Email = fwtypes.StringValue(data.Email, data.JSON.Email)
	m.Name = fwtypes.StringValue(data.Name, data.JSON.Name)
	m.Role = fwtypes.StringValue(data.Role, data.JSON.Role)
	m.AddedAt = fwtypes.TimeValue(data.AddedAt, data.JSON.AddedAt)
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
