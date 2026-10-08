package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apijson"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type WorkspaceModel struct {
	Id            types.String                                                      `tfsdk:"id" apijson:",computed"`
	Name          types.String                                                      `tfsdk:"name" apijson:",computed"`
	CreatedAt     types.String                                                      `tfsdk:"created_at" apijson:",computed"`
	ArchivedAt    types.String                                                      `tfsdk:"archived_at" apijson:",computed"`
	DisplayColor  types.String                                                      `tfsdk:"display_color" apijson:",computed"`
	CompartmentId types.String                                                      `tfsdk:"compartment_id" apijson:",computed"`
	DataResidency supertypes.SingleNestedObjectValueOf[WorkspaceModelDataResidency] `tfsdk:"data_residency" apijson:",computed"`
	ExternalKeyId types.String                                                      `tfsdk:"external_key_id" apijson:",computed"`
	Tags          supertypes.MapValueOf[string]                                     `tfsdk:"tags" apijson:",computed"`
}

func (m *WorkspaceModel) FromAPI(ctx context.Context, data anthropic.Workspace) (diags diag.Diagnostics) {
	err := apijson.DecodeComputed(ctx, data, m)
	if err != nil {
		diags.AddError("Failed to decode Workspace", err.Error())
	}
	return
}

func (m *WorkspaceModel) ToAPIForCreate(ctx context.Context) (*anthropic.OrganizationWorkspaceNewParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := anthropic.OrganizationWorkspaceNewParams{
		Name: m.Name.ValueString(),
	}

	if fwtypes.IsKnown(m.ExternalKeyId) {
		body.ExternalKeyID = anthropic.String(m.ExternalKeyId.ValueString())
	}

	if fwtypes.IsKnown(m.Tags) {
		v := fwdiag.Merge(m.Tags.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		body.Tags = v
	}

	if fwtypes.IsKnown(&m.DataResidency) {
		v := fwdiag.Merge(m.DataResidency.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		vv := fwdiag.Merge(v.ToAPIForCreate(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		body.DataResidency = *vv
	}

	return &body, diags
}

func (m *WorkspaceModel) ToAPIForUpdate(ctx context.Context) (*anthropic.OrganizationWorkspaceUpdateParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := anthropic.OrganizationWorkspaceUpdateParams{
		Name: anthropic.String(m.Name.ValueString()),
	}

	if fwtypes.IsKnown(m.Tags) {
		v := fwdiag.Merge(m.Tags.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		body.Tags = v
	}

	if fwtypes.IsKnown(&m.DataResidency) {
		v := fwdiag.Merge(m.DataResidency.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		vv := fwdiag.Merge(v.ToAPIForUpdate(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		body.DataResidency = *vv
	}

	return &body, diags
}

type WorkspaceModelDataResidency struct {
	AllowedInferenceGeos supertypes.SingleNestedObjectValueOf[WorkspaceModelDataResidencyAllowedInferenceGeos] `tfsdk:"allowed_inference_geos" apijson:"AllowedInferenceGeos,computed"`
	DefaultInferenceGeo  types.String                                                                          `tfsdk:"default_inference_geo" apijson:",computed"`
	WorkspaceGeo         types.String                                                                          `tfsdk:"workspace_geo" apijson:",computed"`
}

func (m *WorkspaceModelDataResidency) FromAPI(ctx context.Context, data anthropic.DataResidency) (diags diag.Diagnostics) {
	m.AllowedInferenceGeos = (func() supertypes.SingleNestedObjectValueOf[WorkspaceModelDataResidencyAllowedInferenceGeos] {
		var mm WorkspaceModelDataResidencyAllowedInferenceGeos
		diags.Append(mm.FromAPI(ctx, data.AllowedInferenceGeos)...)
		return supertypes.NewSingleNestedObjectValueOf(ctx, &mm)
	})()
	m.DefaultInferenceGeo = fwtypes.StringValue(data.DefaultInferenceGeo, data.JSON.DefaultInferenceGeo)
	m.WorkspaceGeo = fwtypes.StringValue(data.WorkspaceGeo, data.JSON.WorkspaceGeo)
	return
}

func (m *WorkspaceModelDataResidency) ToAPIForCreate(ctx context.Context) (*anthropic.DataResidencyCreateConfigParam, diag.Diagnostics) {
	var diags diag.Diagnostics
	var body anthropic.DataResidencyCreateConfigParam
	if fwtypes.IsKnown(m.AllowedInferenceGeos) {
		mm := fwdiag.Merge(m.AllowedInferenceGeos.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		if fwtypes.IsKnown(mm.Values) {
			v := fwdiag.Merge(mm.Values.Get(ctx))(&diags)
			if diags.HasError() {
				return nil, diags
			}

			body.AllowedInferenceGeos.OfGeos = lo.Map(v, func(vv string, _ int) anthropic.AllowedInferenceGeo {
				return anthropic.AllowedInferenceGeo(vv)
			})
		} else if fwtypes.IsKnown(mm.Unrestricted) && mm.Unrestricted.ValueBool() {
			body.AllowedInferenceGeos.OfUnrestricted = constant.ValueOf[constant.Unrestricted]()
		}
	}

	if fwtypes.IsKnown(m.DefaultInferenceGeo) {
		body.DefaultInferenceGeo = anthropic.DataResidencyCreateConfigDefaultInferenceGeo(m.DefaultInferenceGeo.ValueString())
	}

	if fwtypes.IsKnown(m.WorkspaceGeo) {
		body.WorkspaceGeo = anthropic.DataResidencyCreateConfigWorkspaceGeo(m.WorkspaceGeo.ValueString())
	}

	return &body, diags
}

func (m *WorkspaceModelDataResidency) ToAPIForUpdate(ctx context.Context) (*anthropic.DataResidencyUpdateConfigParam, diag.Diagnostics) {
	var diags diag.Diagnostics
	var body anthropic.DataResidencyUpdateConfigParam
	if fwtypes.IsKnown(m.AllowedInferenceGeos) {
		mm := fwdiag.Merge(m.AllowedInferenceGeos.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		if fwtypes.IsKnown(mm.Values) {
			v := fwdiag.Merge(mm.Values.Get(ctx))(&diags)
			if diags.HasError() {
				return nil, diags
			}
			body.AllowedInferenceGeos.OfGeos = lo.Map(v, func(vv string, _ int) anthropic.AllowedInferenceGeo {
				return anthropic.AllowedInferenceGeo(vv)
			})
		} else if fwtypes.IsKnown(mm.Unrestricted) && mm.Unrestricted.ValueBool() {
			body.AllowedInferenceGeos.OfUnrestricted = constant.ValueOf[constant.Unrestricted]()
		}
	}

	if fwtypes.IsKnown(m.DefaultInferenceGeo) {
		body.DefaultInferenceGeo = anthropic.DataResidencyUpdateConfigDefaultInferenceGeo(m.DefaultInferenceGeo.ValueString())
	}

	return &body, diags
}

type WorkspaceModelDataResidencyAllowedInferenceGeos struct {
	Values       supertypes.SetValueOf[string] `tfsdk:"values" apijson:",computed"`
	Unrestricted types.Bool                    `tfsdk:"unrestricted" apijson:",computed"`
}

func (m *WorkspaceModelDataResidencyAllowedInferenceGeos) FromAPI(ctx context.Context, data anthropic.DataResidencyAllowedInferenceGeosUnion) (diags diag.Diagnostics) {
	if data.JSON.OfGeos.Valid() {
		m.Values = supertypes.NewSetValueOfSlice(ctx, lo.Map(data.OfGeos, func(v anthropic.AllowedInferenceGeo, _ int) string {
			return string(v)
		}))
	} else {
		m.Values = supertypes.NewSetValueOfNull[string](ctx)
	}
	m.Unrestricted = types.BoolValue(data.JSON.OfUnrestricted.Valid())
	return
}
