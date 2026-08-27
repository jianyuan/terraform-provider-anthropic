package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type OrganizationInviteModel struct {
	Id        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"` // TODO: rename to invited_at
	ExpiresAt types.String `tfsdk:"expires_at"`
}

func (m *OrganizationInviteModel) Fill(data apiclient.Invite) error {
	m.Id = types.StringValue(data.Id)
	m.Email = types.StringValue(data.Email)
	m.Role = types.StringValue(string(data.Role))
	m.Status = types.StringValue(string(data.Status))
	m.CreatedAt = types.StringValue(data.InvitedAt)
	m.ExpiresAt = types.StringValue(data.ExpiresAt)

	return nil
}
