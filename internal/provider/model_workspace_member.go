package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type WorkspaceMemberModel struct {
	WorkspaceId   types.String `tfsdk:"workspace_id"`
	UserId        types.String `tfsdk:"user_id"`
	WorkspaceRole types.String `tfsdk:"workspace_role"`
}

func (m *WorkspaceMemberModel) Fill(data apiclient.WorkspaceMember) error {
	m.WorkspaceId = types.StringValue(data.WorkspaceId)
	m.UserId = types.StringValue(data.UserId)
	m.WorkspaceRole = types.StringValue(data.WorkspaceRole)

	return nil
}
