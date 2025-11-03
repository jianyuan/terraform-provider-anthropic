package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jianyuan/go-utils/ptr"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type OrganizationInvitesDataSourceModel struct {
	Invites []OrganizationInviteDataSourceModel `tfsdk:"invites"`
}

type OrganizationInviteDataSourceModel struct {
	Id        string `tfsdk:"id"`
	Email     string `tfsdk:"email"`
	Role      string `tfsdk:"role"`
	Status    string `tfsdk:"status"`
	CreatedAt string `tfsdk:"created_at"`
	ExpiresAt string `tfsdk:"expires_at"`
}

func (m *OrganizationInvitesDataSourceModel) Fill(invites []apiclient.Invite) error {
	m.Invites = make([]OrganizationInviteDataSourceModel, len(invites))
	for i, inv := range invites {
		m.Invites[i] = OrganizationInviteDataSourceModel{
			Id:        inv.Id,
			Email:     inv.Email,
			Role:      inv.Role,
			Status:    inv.Status,
			CreatedAt: inv.CreatedAt,
			ExpiresAt: inv.ExpiresAt,
		}
	}

	return nil
}

func NewOrganizationInvitesDataSource() datasource.DataSource {
	return &OrganizationInvitesDataSource{}
}

var _ datasource.DataSource = &OrganizationInvitesDataSource{}

type OrganizationInvitesDataSource struct {
	baseDataSource
}

func (d *OrganizationInvitesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_invites"
}

func (d *OrganizationInvitesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all pending invites in the Organization.",

		Attributes: map[string]schema.Attribute{
			"invites": schema.SetNestedAttribute{
				MarkdownDescription: "List of pending organization invites.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique identifier for the invite.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "Email address of the person being invited.",
							Computed:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "Role to assign to the invited user.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "Current status of the invite (e.g., pending, accepted, expired).",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "RFC 3339 datetime string indicating when the invite was created.",
							Computed:            true,
						},
						"expires_at": schema.StringAttribute{
							MarkdownDescription: "RFC 3339 datetime string indicating when the invite expires.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationInvitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationInvitesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var invites []apiclient.Invite
	params := &apiclient.ListInvitesParams{
		Limit: ptr.Ptr(100),
	}

	for {
		httpResp, err := d.client.ListInvitesWithResponse(
			ctx,
			params,
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read invites, got error: %s", err))
			return
		}

		if httpResp.StatusCode() != 200 {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read invites, got status code: %d", httpResp.StatusCode()))
			return
		}

		if httpResp.JSON200 == nil {
			resp.Diagnostics.AddError("Client Error", "Unable to read invites, got empty response body")
			return
		}

		invites = append(invites, httpResp.JSON200.Data...)

		if !httpResp.JSON200.HasMore || httpResp.JSON200.LastId == nil {
			break
		}

		params.AfterId = httpResp.JSON200.LastId
	}

	if err := data.Fill(invites); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
