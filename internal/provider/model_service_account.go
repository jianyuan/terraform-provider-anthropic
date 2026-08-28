package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
)

type ServiceAccountModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	OrganizationRole  types.String `tfsdk:"organization_role"`
	ArchivedAt        types.String `tfsdk:"archived_at"`
	ArchivedByActorId types.String `tfsdk:"archived_by_actor_id"`
	CreatedAt         types.String `tfsdk:"created_at"`
	CreatedByActorId  types.String `tfsdk:"created_by_actor_id"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	UpdatedByActorId  types.String `tfsdk:"updated_by_actor_id"`
}

func (m *ServiceAccountModel) FromAPI(ctx context.Context, data apiclient.ServiceAccount) (diags diag.Diagnostics) {
	m.Id = types.StringValue(data.Id)
	m.Name = types.StringValue(data.Name)
	m.Description = fwtypes.NullableStringValue(data.Description)
	m.OrganizationRole = types.StringValue(string(data.OrganizationRole))
	m.ArchivedAt = fwtypes.NullableStringValue(data.ArchivedAt)
	m.ArchivedByActorId = fwtypes.NullableStringValue(data.ArchivedByActorId)
	m.CreatedAt = types.StringValue(data.CreatedAt)
	m.CreatedByActorId = fwtypes.NullableStringValue(data.CreatedByActorId)
	m.UpdatedAt = types.StringValue(data.UpdatedAt)
	m.UpdatedByActorId = fwtypes.NullableStringValue(data.UpdatedByActorId)
	return
}

func (m *ServiceAccountModel) ToAPIForCreate(ctx context.Context) (*apiclient.CreateServiceAccountJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := apiclient.CreateServiceAccountJSONRequestBody{
		Name: m.Name.ValueString(),
	}

	if fwtypes.IsKnown(m.Description) {
		body.Description.Set(m.Description.ValueString())
	}

	if fwtypes.IsKnown(m.OrganizationRole) {
		body.OrganizationRole = new(apiclient.CreateServiceAccountRequestOrganizationRole(m.OrganizationRole.ValueString()))
	}

	return &body, diags
}

func (m *ServiceAccountModel) ToAPIForUpdate(ctx context.Context) (*apiclient.UpdateServiceAccountJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := apiclient.UpdateServiceAccountJSONRequestBody{}

	if fwtypes.IsKnown(m.Description) {
		body.Description.Set(m.Description.ValueString())
	}

	if fwtypes.IsKnown(m.OrganizationRole) {
		body.OrganizationRole.Set(apiclient.UpdateServiceAccountRequestOrganizationRole(m.OrganizationRole.ValueString()))
	}

	return &body, diags
}
