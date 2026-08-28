package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type WorkspaceMemberModel struct {
	WorkspaceId   types.String `tfsdk:"workspace_id"`
	UserId        types.String `tfsdk:"user_id"`
	WorkspaceRole types.String `tfsdk:"workspace_role"`
}

func (m *WorkspaceMemberModel) Fill(ctx context.Context, member apiclient.WorkspaceMember) (diags diag.Diagnostics) {
	m.WorkspaceId = types.StringValue(member.WorkspaceId)
	m.UserId = types.StringValue(member.UserId)
	m.WorkspaceRole = types.StringValue(string(member.WorkspaceRole))
	return
}
