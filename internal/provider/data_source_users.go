package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type UsersDataSourceModel struct {
	Users supertypes.SetNestedObjectValueOf[UserDataSourceModel] `tfsdk:"users"`
}

func (m *UsersDataSourceModel) FromAPI(ctx context.Context, users []apiclient.User) (diags diag.Diagnostics) {
	m.Users = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(users, func(user apiclient.User, _ int) UserDataSourceModel {
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
	var data UsersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var users []apiclient.User
	params := &apiclient.ListUsersParams{
		Limit: new(int64(100)),
	}

	for {
		page := fwdiag.Merge(apiclient.ReadJSON200(d.client.ListUsersWithResponse(
			ctx,
			params,
			d.WithApiKeyRequestEditorFn(),
		)))(&resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		users = append(users, page.Data...)

		if v, err := page.LastId.Get(); err == nil {
			params.AfterId = new(v)
		} else {
			break
		}
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, users)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
