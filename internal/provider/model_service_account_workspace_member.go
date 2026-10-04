package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
)

type ServiceAccountWorkspaceMemberModel struct {
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`
	WorkspaceRole    types.String `tfsdk:"workspace_role"`
	CreatedByActorId types.String `tfsdk:"created_by_actor_id"`
	Implicit         types.Bool   `tfsdk:"implicit"`
}

func (m *ServiceAccountWorkspaceMemberModel) FromAPI(ctx context.Context, data anthropic.ServiceAccountWorkspaceMember) (diags diag.Diagnostics) {
	m.WorkspaceId = fwtypes.StringValue(data.WorkspaceID, data.JSON.WorkspaceID)
	m.ServiceAccountId = fwtypes.StringValue(data.ServiceAccountID, data.JSON.ServiceAccountID)
	m.WorkspaceRole = fwtypes.StringValue(data.WorkspaceRole, data.JSON.WorkspaceRole)
	m.CreatedByActorId = fwtypes.StringValue(data.CreatedByActorID, data.JSON.CreatedByActorID)
	m.Implicit = fwtypes.BoolValue(data.Implicit, data.JSON.Implicit)
	return
}

func (m *ServiceAccountWorkspaceMemberModel) ToAPIForCreate(ctx context.Context) (*anthropic.OrganizationWorkspaceServiceAccountAddParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	return &anthropic.OrganizationWorkspaceServiceAccountAddParams{
		ServiceAccountID: m.ServiceAccountId.ValueString(),
		WorkspaceRole:    anthropic.NoBillingWorkspaceRole(m.WorkspaceRole.ValueString()),
	}, diags
}

func (m *ServiceAccountWorkspaceMemberModel) ToAPIForUpdate(ctx context.Context) (*anthropic.OrganizationWorkspaceServiceAccountUpdateParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	return &anthropic.OrganizationWorkspaceServiceAccountUpdateParams{
		WorkspaceID:   m.WorkspaceId.ValueString(),
		WorkspaceRole: anthropic.NoBillingWorkspaceRole(m.WorkspaceRole.ValueString()),
	}, diags
}
