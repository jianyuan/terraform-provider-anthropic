package provider

import (
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

func (m *WorkspaceModel) Fill(w apiclient.Workspace) error {
	m.Id = types.StringValue(w.Id)
	m.Name = types.StringValue(w.Name)
	m.CreatedAt = types.StringValue(w.CreatedAt)
	m.ArchivedAt = types.StringPointerValue(w.ArchivedAt)
	m.DisplayColor = types.StringValue(w.DisplayColor)

	return nil
}
