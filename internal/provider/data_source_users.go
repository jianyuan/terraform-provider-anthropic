package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jianyuan/go-utils/ptr"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type UsersDataSourceModel struct {
	Users []UserDataSourceModel `tfsdk:"users"`
}

func (m *UsersDataSourceModel) Fill(users []apiclient.User) error {
	m.Users = make([]UserDataSourceModel, len(users))
	for i, u := range users {
		if err := m.Users[i].Fill(u); err != nil {
			return err
		}
	}

	return nil
}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

var _ datasource.DataSource = &UsersDataSource{}

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
		Limit: ptr.Ptr(100),
	}

	for {
		httpResp, err := d.client.ListUsersWithResponse(
			ctx,
			params,
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got status code: %d", httpResp.StatusCode()))
			return
		}

		if httpResp.JSON200 == nil {
			resp.Diagnostics.AddError("Client Error", "Unable to read, got empty response body")
			return
		}

		users = append(users, httpResp.JSON200.Data...)

		if !httpResp.JSON200.HasMore || httpResp.JSON200.LastId == nil {
			break
		}

		params.AfterId = httpResp.JSON200.LastId
	}

	if err := data.Fill(users); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
