package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (m *OrganizationInviteModel) FromAPI(ctx context.Context, data anthropic.OrganizationInvite) (diags diag.Diagnostics) {
	m.Id = fwtypes.StringValue(data.ID, data.JSON.ID)
	m.Email = fwtypes.StringValue(data.Email, data.JSON.Email)
	m.Role = fwtypes.StringValue(data.Role, data.JSON.Role)
	if len(data.RBACGroupIDs) == 0 {
		m.RbacGroupIds = supertypes.NewSetValueOfNull[string](ctx)
	} else {
		m.RbacGroupIds = supertypes.NewSetValueOfSlice(ctx, data.RBACGroupIDs)
	}
	m.Status = fwtypes.StringValue(data.Status, data.JSON.Status)
	m.AcceptedAt = fwtypes.TimeValue(data.AcceptedAt, data.JSON.AcceptedAt)
	m.InvitedAt = fwtypes.TimeValue(data.InvitedAt, data.JSON.InvitedAt)
	m.ExpiresAt = fwtypes.TimeValue(data.ExpiresAt, data.JSON.ExpiresAt)
	return
}
