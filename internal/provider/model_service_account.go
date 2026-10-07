package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (m *ServiceAccountModel) FromAPI(ctx context.Context, data anthropic.ServiceAccount) (diags diag.Diagnostics) {
	m.Id = fwtypes.StringValue(data.ID, data.JSON.ID)
	m.Name = fwtypes.StringValue(data.Name, data.JSON.Name)
	m.Description = fwtypes.StringValue(data.Description, data.JSON.Description)
	m.OrganizationRole = fwtypes.StringValue(data.OrganizationRole, data.JSON.OrganizationRole)
	m.ArchivedAt = fwtypes.TimeValue(data.ArchivedAt, data.JSON.ArchivedAt)
	m.ArchivedByActorId = fwtypes.StringValue(data.ArchivedByActorID, data.JSON.ArchivedByActorID)
	m.CreatedAt = fwtypes.TimeValue(data.CreatedAt, data.JSON.CreatedAt)
	m.CreatedByActorId = fwtypes.StringValue(data.CreatedByActorID, data.JSON.CreatedByActorID)
	m.UpdatedAt = fwtypes.TimeValue(data.UpdatedAt, data.JSON.UpdatedAt)
	m.UpdatedByActorId = fwtypes.StringValue(data.UpdatedByActorID, data.JSON.UpdatedByActorID)
	return
}

func (m *ServiceAccountModel) ToAPIForCreate(ctx context.Context) (*anthropic.OrganizationServiceAccountNewParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := anthropic.OrganizationServiceAccountNewParams{
		Name: m.Name.ValueString(),
	}

	if fwtypes.IsKnown(m.Description) {
		body.Description = anthropic.String(m.Description.ValueString())
	}

	if fwtypes.IsKnown(m.OrganizationRole) {
		body.OrganizationRole = anthropic.OrganizationServiceAccountNewParamsOrganizationRole(m.OrganizationRole.ValueString())
	}

	return &body, diags
}

func (m *ServiceAccountModel) ToAPIForUpdate(ctx context.Context) (*anthropic.OrganizationServiceAccountUpdateParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := anthropic.OrganizationServiceAccountUpdateParams{}

	if fwtypes.IsKnown(m.Description) {
		body.Description = anthropic.String(m.Description.ValueString())
	}

	if fwtypes.IsKnown(m.OrganizationRole) {
		body.OrganizationRole = anthropic.OrganizationServiceAccountUpdateParamsOrganizationRole(m.OrganizationRole.ValueString())
	}

	return &body, diags
}
