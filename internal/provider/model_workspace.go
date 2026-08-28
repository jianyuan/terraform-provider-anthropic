package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
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

func (m *WorkspaceModel) FromAPI(ctx context.Context, data apiclient.Workspace) (diags diag.Diagnostics) {
	m.Id = types.StringValue(data.Id)
	m.Name = types.StringValue(data.Name)
	m.CreatedAt = types.StringValue(data.CreatedAt)
	m.ArchivedAt = fwtypes.NullableStringValue(data.ArchivedAt)
	m.DisplayColor = types.StringValue(data.DisplayColor)
	m.CompartmentId = types.StringValue(data.CompartmentId)
	m.DataResidency = (func() supertypes.SingleNestedObjectValueOf[WorkspaceModelDataResidency] {
		var mm WorkspaceModelDataResidency
		diags.Append(mm.FromAPI(ctx, data.DataResidency)...)
		return supertypes.NewSingleNestedObjectValueOf(ctx, &mm)
	})()
	m.ExternalKeyId = fwtypes.NullableStringValue(data.ExternalKeyId)
	m.Tags = fwdiag.Merge(supertypes.NewMapValueOfMap(ctx, data.Tags))(&diags)
	return
}

func (m *WorkspaceModel) ToAPIForCreate(ctx context.Context) (*apiclient.CreateWorkspaceJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := apiclient.CreateWorkspaceJSONRequestBody{
		Name: m.Name.ValueString(),
	}

	if fwtypes.IsKnown(m.ExternalKeyId) {
		body.ExternalKeyId.Set(m.ExternalKeyId.ValueString())
	}

	if fwtypes.IsKnown(m.Tags) {
		v := fwdiag.Merge(m.Tags.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		body.Tags.Set(v)
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

		body.DataResidency.Set(*vv)
	}

	return &body, diags
}

func (m *WorkspaceModel) ToAPIForUpdate(ctx context.Context) (*apiclient.UpdateWorkspaceJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := apiclient.UpdateWorkspaceJSONRequestBody{
		Name: m.Name.ValueStringPointer(),
	}

	if fwtypes.IsKnown(m.Tags) {
		v := fwdiag.Merge(m.Tags.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		body.Tags.Set(v)
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

		body.DataResidency.Set(*vv)
	}

	return &body, diags
}

type WorkspaceModelDataResidency struct {
	AllowedInferenceGeos supertypes.SingleNestedObjectValueOf[WorkspaceModelDataResidencyAllowedInferenceGeos] `tfsdk:"allowed_inference_geos"`
	DefaultInferenceGeo  types.String                                                                          `tfsdk:"default_inference_geo"`
	WorkspaceGeo         types.String                                                                          `tfsdk:"workspace_geo"`
}

func (m *WorkspaceModelDataResidency) FromAPI(ctx context.Context, data apiclient.Workspace_DataResidency) (diags diag.Diagnostics) {
	m.AllowedInferenceGeos = (func() supertypes.SingleNestedObjectValueOf[WorkspaceModelDataResidencyAllowedInferenceGeos] {
		var mm WorkspaceModelDataResidencyAllowedInferenceGeos
		diags.Append(mm.FromAPI(ctx, data)...)
		return supertypes.NewSingleNestedObjectValueOf(ctx, &mm)
	})()
	m.DefaultInferenceGeo = types.StringValue(data.DefaultInferenceGeo)
	m.WorkspaceGeo = types.StringValue(data.WorkspaceGeo)
	return
}

func (m *WorkspaceModelDataResidency) ToAPIForCreate(ctx context.Context) (*apiclient.CreateWorkspaceRequest_DataResidency, diag.Diagnostics) {
	var diags diag.Diagnostics
	var body apiclient.CreateWorkspaceRequest_DataResidency
	if fwtypes.IsKnown(m.AllowedInferenceGeos) {
		body.AllowedInferenceGeos = &apiclient.CreateWorkspaceRequest_DataResidency_AllowedInferenceGeos{}

		mm := fwdiag.Merge(m.AllowedInferenceGeos.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		if fwtypes.IsKnown(mm.Values) {
			v := fwdiag.Merge(mm.Values.Get(ctx))(&diags)
			if diags.HasError() {
				return nil, diags
			}
			if err := body.AllowedInferenceGeos.FromCreateWorkspaceRequestDataResidencyAllowedInferenceGeos0(v); err != nil {
				diags.AddError("Failed to convert AllowedInferenceGeos", err.Error())
				return nil, diags
			}
		} else if fwtypes.IsKnown(mm.Unrestricted) && mm.Unrestricted.ValueBool() {
			if err := body.AllowedInferenceGeos.FromCreateWorkspaceRequestDataResidencyAllowedInferenceGeos1(apiclient.CreateWorkspaceRequestDataResidencyAllowedInferenceGeos1Unrestricted); err != nil {
				diags.AddError("Failed to convert AllowedInferenceGeos", err.Error())
				return nil, diags
			}
		}
	}

	if fwtypes.IsKnown(m.DefaultInferenceGeo) {
		body.DefaultInferenceGeo.Set(apiclient.CreateWorkspaceRequestDataResidencyDefaultInferenceGeo(m.DefaultInferenceGeo.ValueString()))
	}

	if fwtypes.IsKnown(m.WorkspaceGeo) {
		body.WorkspaceGeo.Set(apiclient.CreateWorkspaceRequestDataResidencyWorkspaceGeo(m.WorkspaceGeo.ValueString()))
	}

	return &body, diags
}

func (m *WorkspaceModelDataResidency) ToAPIForUpdate(ctx context.Context) (*apiclient.UpdateWorkspaceRequest_DataResidency, diag.Diagnostics) {
	var diags diag.Diagnostics
	var body apiclient.UpdateWorkspaceRequest_DataResidency
	if fwtypes.IsKnown(m.AllowedInferenceGeos) {
		body.AllowedInferenceGeos = &apiclient.UpdateWorkspaceRequest_DataResidency_AllowedInferenceGeos{}

		mm := fwdiag.Merge(m.AllowedInferenceGeos.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		if fwtypes.IsKnown(mm.Values) {
			v := fwdiag.Merge(mm.Values.Get(ctx))(&diags)
			if diags.HasError() {
				return nil, diags
			}
			if err := body.AllowedInferenceGeos.FromUpdateWorkspaceRequestDataResidencyAllowedInferenceGeos0(v); err != nil {
				diags.AddError("Failed to convert AllowedInferenceGeos", err.Error())
				return nil, diags
			}
		} else if fwtypes.IsKnown(mm.Unrestricted) && mm.Unrestricted.ValueBool() {
			if err := body.AllowedInferenceGeos.FromUpdateWorkspaceRequestDataResidencyAllowedInferenceGeos1(apiclient.UpdateWorkspaceRequestDataResidencyAllowedInferenceGeos1Unrestricted); err != nil {
				diags.AddError("Failed to convert AllowedInferenceGeos", err.Error())
				return nil, diags
			}
		}
	}

	if fwtypes.IsKnown(m.DefaultInferenceGeo) {
		body.DefaultInferenceGeo.Set(apiclient.UpdateWorkspaceRequestDataResidencyDefaultInferenceGeo(m.DefaultInferenceGeo.ValueString()))
	}

	return &body, diags
}

type WorkspaceModelDataResidencyAllowedInferenceGeos struct {
	Values       supertypes.SetValueOf[string] `tfsdk:"values"`
	Unrestricted types.Bool                    `tfsdk:"unrestricted"`
}

func (m *WorkspaceModelDataResidencyAllowedInferenceGeos) FromAPI(ctx context.Context, data apiclient.Workspace_DataResidency) (diags diag.Diagnostics) {
	if v, err := data.AllowedInferenceGeos.AsWorkspaceDataResidencyAllowedInferenceGeos0(); err == nil {
		fmt.Println(v)
		m.Values = supertypes.NewSetValueOfSlice(ctx, v)
		m.Unrestricted = types.BoolValue(false)
	} else if v, err := data.AllowedInferenceGeos.AsWorkspaceDataResidencyAllowedInferenceGeos1(); err == nil {
		m.Values = supertypes.NewSetValueOfNull[string](ctx)
		m.Unrestricted = types.BoolValue(v == apiclient.WorkspaceDataResidencyAllowedInferenceGeos1Unrestricted)
	}
	return
}
