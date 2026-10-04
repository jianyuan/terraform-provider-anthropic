package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type UsersDataSourceModel struct {
	Users supertypes.SetNestedObjectValueOf[UserDataSourceModel] `tfsdk:"users"`
}

func (m *UsersDataSourceModel) FromAPI(ctx context.Context, users []anthropic.OrganizationUser) (diags diag.Diagnostics) {
	m.Users = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(users, func(user anthropic.OrganizationUser, _ int) UserDataSourceModel {
		var mm UserDataSourceModel
		diags.Append(mm.FromAPI(ctx, user)...)
		return mm
	}))
	return
}

var _ datasource.DataSource = &UsersDataSource{}
var _ datasource.DataSourceWithConfigure = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	baseDataSource
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all users in the Organization.",

		Attributes: map[string]schema.Attribute{
			"users": schema.SetNestedAttribute{
				MarkdownDescription: "List of users.",
				Computed:            true,
				CustomType:          supertypes.NewSetNestedObjectTypeOf[UserDataSourceModel](ctx),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "ID of the User.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	fwdatasource.List(ctx, func(_ *UsersDataSourceModel) fwdatasource.AutoPager[anthropic.OrganizationUser] {
		return d.apiKeyClient.Organization.Users.ListAutoPaging(ctx, anthropic.OrganizationUserListParams{})
	}, req, resp)
}
