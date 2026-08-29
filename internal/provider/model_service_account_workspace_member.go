package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
)

type ServiceAccountWorkspaceMemberModel struct {
	WorkspaceId      types.String `tfsdk:"workspace_id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`
	WorkspaceRole    types.String `tfsdk:"workspace_role"`
	CreatedByActorId types.String `tfsdk:"created_by_actor_id"`
	Implicit         types.Bool   `tfsdk:"implicit"`
}

func (m *ServiceAccountWorkspaceMemberModel) FromCreateAPI(ctx context.Context, data apiclient.CreateServiceAccountWorkspaceMember200JSONResponseBody) (diags diag.Diagnostics) {
	m.WorkspaceId = types.StringValue(data.WorkspaceId)
	m.ServiceAccountId = types.StringValue(data.ServiceAccountId)
	m.WorkspaceRole = types.StringValue(string(data.WorkspaceRole))
	m.CreatedByActorId = fwtypes.NullableStringValue(data.CreatedByActorId)
	m.Implicit = fwtypes.NullableBoolValue(data.Implicit)
	return
}

func (m *ServiceAccountWorkspaceMemberModel) FromReadAPI(ctx context.Context, data apiclient.GetServiceAccountWorkspaceMember200JSONResponseBody) (diags diag.Diagnostics) {
	m.WorkspaceId = types.StringValue(data.WorkspaceId)
	m.ServiceAccountId = types.StringValue(data.ServiceAccountId)
	m.WorkspaceRole = types.StringValue(string(data.WorkspaceRole))
	m.CreatedByActorId = fwtypes.NullableStringValue(data.CreatedByActorId)
	m.Implicit = fwtypes.NullableBoolValue(data.Implicit)
	return
}

func (m *ServiceAccountWorkspaceMemberModel) FromUpdateAPI(ctx context.Context, data apiclient.UpdateServiceAccountWorkspaceMember200JSONResponseBody) (diags diag.Diagnostics) {
	m.WorkspaceId = types.StringValue(data.WorkspaceId)
	m.ServiceAccountId = types.StringValue(data.ServiceAccountId)
	m.WorkspaceRole = types.StringValue(string(data.WorkspaceRole))
	m.CreatedByActorId = fwtypes.NullableStringValue(data.CreatedByActorId)
	m.Implicit = fwtypes.NullableBoolValue(data.Implicit)
	return
}

func (m *ServiceAccountWorkspaceMemberModel) ToAPIForCreate(ctx context.Context) (*apiclient.CreateServiceAccountWorkspaceMemberJSONRequestBody, diag.Diagnostics) {
	return &apiclient.CreateServiceAccountWorkspaceMemberJSONRequestBody{
		ServiceAccountId: m.ServiceAccountId.ValueString(),
		WorkspaceRole:    apiclient.CreateServiceAccountWorkspaceMemberRequestWorkspaceRole(m.WorkspaceRole.ValueString()),
	}, nil
}

func (m *ServiceAccountWorkspaceMemberModel) ToAPIForUpdate(ctx context.Context) (*apiclient.UpdateServiceAccountWorkspaceMemberJSONRequestBody, diag.Diagnostics) {
	return &apiclient.UpdateServiceAccountWorkspaceMemberJSONRequestBody{
		WorkspaceRole: apiclient.UpdateServiceAccountWorkspaceMemberRequestWorkspaceRole(m.WorkspaceRole.ValueString()),
	}, nil
}
