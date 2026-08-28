package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

type UserDataSourceModel struct {
	Id      types.String `tfsdk:"id"`
	Email   types.String `tfsdk:"email"`
	Name    types.String `tfsdk:"name"`
	Role    types.String `tfsdk:"role"`
	AddedAt types.String `tfsdk:"added_at"`
}

func (m *UserDataSourceModel) FromAPI(ctx context.Context, data apiclient.User) (diags diag.Diagnostics) {
	m.Id = types.StringValue(data.Id)
	m.Email = types.StringValue(data.Email)
	m.Name = types.StringValue(data.Name)
	m.Role = types.StringValue(string(data.Role))
	m.AddedAt = types.StringValue(data.AddedAt)
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
	var data UserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := fwdiag.Merge(apiclient.ReadJSON200(d.client.GetUserWithResponse(
		ctx,
		data.Id.ValueString(),
		d.WithApiKeyRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *user)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
