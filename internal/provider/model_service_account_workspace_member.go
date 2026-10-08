package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apijson"
)

type ServiceAccountWorkspaceMemberModel struct {
	WorkspaceId      types.String `tfsdk:"workspace_id" apijson:",computed"`
	ServiceAccountId types.String `tfsdk:"service_account_id" apijson:",computed"`
	WorkspaceRole    types.String `tfsdk:"workspace_role" apijson:",computed"`
	CreatedByActorId types.String `tfsdk:"created_by_actor_id" apijson:",computed"`
	Implicit         types.Bool   `tfsdk:"implicit" apijson:",computed"`
}

func (m *ServiceAccountWorkspaceMemberModel) FromAPI(ctx context.Context, data anthropic.ServiceAccountWorkspaceMember) (diags diag.Diagnostics) {
	err := apijson.DecodeComputed(ctx, data, m)
	if err != nil {
		diags.AddError("Failed to decode ServiceAccountWorkspaceMember", err.Error())
	}
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
