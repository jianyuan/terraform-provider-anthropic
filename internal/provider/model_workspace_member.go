package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
)

type WorkspaceMemberModel struct {
	WorkspaceId   types.String `tfsdk:"workspace_id"`
	UserId        types.String `tfsdk:"user_id"`
	WorkspaceRole types.String `tfsdk:"workspace_role"`
}

func (m *WorkspaceMemberModel) FromAPI(ctx context.Context, member anthropic.WorkspaceMember) (diags diag.Diagnostics) {
	m.WorkspaceId = fwtypes.StringValue(member.WorkspaceID, member.JSON.WorkspaceID)
	m.UserId = fwtypes.StringValue(member.UserID, member.JSON.UserID)
	m.WorkspaceRole = fwtypes.StringValue(member.WorkspaceRole, member.JSON.WorkspaceRole)
	return
}

func (m *WorkspaceMemberModel) ToAPIForCreate(ctx context.Context) (*anthropic.OrganizationWorkspaceMemberAddParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	return &anthropic.OrganizationWorkspaceMemberAddParams{
		UserID:        m.UserId.ValueString(),
		WorkspaceRole: anthropic.NoBillingWorkspaceRole(m.WorkspaceRole.ValueString()),
	}, diags
}

func (m *WorkspaceMemberModel) ToAPIForUpdate(ctx context.Context) (*anthropic.OrganizationWorkspaceMemberUpdateParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	return &anthropic.OrganizationWorkspaceMemberUpdateParams{
		WorkspaceRole: anthropic.WorkspaceRole(m.WorkspaceRole.ValueString()),
		WorkspaceID:   m.WorkspaceId.ValueString(),
	}, diags
}
