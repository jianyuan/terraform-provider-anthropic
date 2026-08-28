package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type WorkspaceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	CreatedAt    types.String `tfsdk:"created_at"`
	ArchivedAt   types.String `tfsdk:"archived_at"`
	DisplayColor types.String `tfsdk:"display_color"`
}

func (m *WorkspaceModel) Fill(ctx context.Context, workspace apiclient.Workspace) (diags diag.Diagnostics) {
	m.Id = types.StringValue(workspace.Id)
	m.Name = types.StringValue(workspace.Name)
	m.CreatedAt = types.StringValue(workspace.CreatedAt)
	if v, err := workspace.ArchivedAt.Get(); err == nil {
		m.ArchivedAt = types.StringValue(v)
	} else {
		m.ArchivedAt = types.StringNull()
	}
	m.DisplayColor = types.StringValue(workspace.DisplayColor)

	return
}
