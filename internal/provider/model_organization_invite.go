package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

type OrganizationInviteModel struct {
	Id           types.String                  `tfsdk:"id"`
	Email        types.String                  `tfsdk:"email"`
	Role         types.String                  `tfsdk:"role"`
	RbacGroupIds supertypes.SetValueOf[string] `tfsdk:"rbac_group_ids"`
	Status       types.String                  `tfsdk:"status"`
	AcceptedAt   types.String                  `tfsdk:"accepted_at"`
	InvitedAt    types.String                  `tfsdk:"invited_at"`
	ExpiresAt    types.String                  `tfsdk:"expires_at"`
}

func (m *OrganizationInviteModel) FromAPI(ctx context.Context, data apiclient.Invite) (diags diag.Diagnostics) {
	m.Id = types.StringValue(data.Id)
	m.Email = types.StringValue(data.Email)
	m.Role = types.StringValue(string(data.Role))
	if len(data.RbacGroupIds) == 0 {
		m.RbacGroupIds = supertypes.NewSetValueOfNull[string](ctx)
	} else {
		m.RbacGroupIds = supertypes.NewSetValueOfSlice(ctx, data.RbacGroupIds)
	}
	m.Status = types.StringValue(string(data.Status))
	m.AcceptedAt = fwtypes.NullableStringValue(data.AcceptedAt)
	m.InvitedAt = types.StringValue(data.InvitedAt)
	m.ExpiresAt = types.StringValue(data.ExpiresAt)
	return
}
