package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

type WorkspaceModel struct {
	Id            types.String                                                      `tfsdk:"id"`
	Name          types.String                                                      `tfsdk:"name"`
	CreatedAt     types.String                                                      `tfsdk:"created_at"`
	ArchivedAt    types.String                                                      `tfsdk:"archived_at"`
	DisplayColor  types.String                                                      `tfsdk:"display_color"`
	CompartmentId types.String                                                      `tfsdk:"compartment_id"`
	DataResidency supertypes.SingleNestedObjectValueOf[WorkspaceModelDataResidency] `tfsdk:"data_residency"`
	ExternalKeyId types.String                                                      `tfsdk:"external_key_id"`
	Tags          supertypes.MapValueOf[string]                                     `tfsdk:"tags"`
}

func (m *WorkspaceModel) Fill(ctx context.Context, data apiclient.Workspace) (diags diag.Diagnostics) {
	m.Id = types.StringValue(data.Id)
	m.Name = types.StringValue(data.Name)
	m.CreatedAt = types.StringValue(data.CreatedAt)
	if v, err := data.ArchivedAt.Get(); err == nil {
		m.ArchivedAt = types.StringValue(v)
	} else {
		m.ArchivedAt = types.StringNull()
	}
	m.DisplayColor = types.StringValue(data.DisplayColor)
	m.CompartmentId = types.StringValue(data.CompartmentId)
	m.DataResidency = (func() supertypes.SingleNestedObjectValueOf[WorkspaceModelDataResidency] {
		var mm WorkspaceModelDataResidency
		diags.Append(mm.Fill(ctx, data.DataResidency)...)
		return supertypes.NewSingleNestedObjectValueOf(ctx, &mm)
	})()
	if v, err := data.ExternalKeyId.Get(); err == nil {
		m.ExternalKeyId = types.StringValue(v)
	} else {
		m.ExternalKeyId = types.StringNull()
	}
	m.Tags = fwdiag.Merge(supertypes.NewMapValueOfMap(ctx, data.Tags))(&diags)
	return
}

type WorkspaceModelDataResidency struct {
	AllowedInferenceGeos             supertypes.SetValueOf[string] `tfsdk:"allowed_inference_geos"`
	AllowedInferenceGeosUnrestricted types.Bool                    `tfsdk:"allowed_inference_geos_unrestricted"`
	DefaultInferenceGeo              types.String                  `tfsdk:"default_inference_geo"`
	WorkspaceGeo                     types.String                  `tfsdk:"workspace_geo"`
}

func (m *WorkspaceModelDataResidency) Fill(ctx context.Context, data apiclient.Workspace_DataResidency) (diags diag.Diagnostics) {
	if v, err := data.AllowedInferenceGeos.AsWorkspaceDataResidencyAllowedInferenceGeos0(); err == nil {
		m.AllowedInferenceGeos = supertypes.NewSetValueOfSlice(ctx, v)
		m.AllowedInferenceGeosUnrestricted = types.BoolNull()
	} else if v, err := data.AllowedInferenceGeos.AsWorkspaceDataResidencyAllowedInferenceGeos1(); err == nil {
		m.AllowedInferenceGeos = supertypes.NewSetValueOfNull[string](ctx)
		m.AllowedInferenceGeosUnrestricted = types.BoolValue(v == apiclient.WorkspaceDataResidencyAllowedInferenceGeos1Unrestricted)
	}
	m.DefaultInferenceGeo = types.StringValue(data.DefaultInferenceGeo)
	m.WorkspaceGeo = types.StringValue(data.WorkspaceGeo)
	return
}
