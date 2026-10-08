package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apijson"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
)

type ServiceAccountModel struct {
	Id                types.String `tfsdk:"id" apijson:",computed"`
	Name              types.String `tfsdk:"name" apijson:",computed"`
	Description       types.String `tfsdk:"description" apijson:",computed"`
	OrganizationRole  types.String `tfsdk:"organization_role" apijson:",computed"`
	ArchivedAt        types.String `tfsdk:"archived_at" apijson:",computed"`
	ArchivedByActorId types.String `tfsdk:"archived_by_actor_id" apijson:",computed"`
	CreatedAt         types.String `tfsdk:"created_at" apijson:",computed"`
	CreatedByActorId  types.String `tfsdk:"created_by_actor_id" apijson:",computed"`
	UpdatedAt         types.String `tfsdk:"updated_at" apijson:",computed"`
	UpdatedByActorId  types.String `tfsdk:"updated_by_actor_id" apijson:",computed"`
}

func (m *ServiceAccountModel) FromAPI(ctx context.Context, data anthropic.ServiceAccount) (diags diag.Diagnostics) {
	err := apijson.DecodeComputed(ctx, data, m)
	if err != nil {
		diags.AddError("Failed to decode ServiceAccount", err.Error())
	}
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
