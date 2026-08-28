package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

var _ resource.Resource = &WorkspaceResource{}
var _ resource.ResourceWithConfigure = &WorkspaceResource{}
var _ resource.ResourceWithImportState = &WorkspaceResource{}

func NewWorkspaceResource() resource.Resource {
	return &WorkspaceResource{}
}

type WorkspaceResource struct {
	baseResource
}

func (r *WorkspaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *WorkspaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = workspaceSchema().GetResource(ctx)
}

func (r *WorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkspaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.CreateWorkspaceJSONRequestBody{
		Name: data.Name.ValueString(),
	}
	if !data.ExternalKeyId.IsNull() && !data.ExternalKeyId.IsUnknown() {
		body.ExternalKeyId.Set(data.ExternalKeyId.ValueString())
	} else {
		body.ExternalKeyId.SetUnspecified()
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		body.Tags.Set(fwdiag.Merge(data.Tags.Get(ctx))(&resp.Diagnostics))
	} else {
		body.Tags.SetUnspecified()
	}

	if !data.DataResidency.IsNull() && !data.DataResidency.IsUnknown() {
		dr := fwdiag.Merge(data.DataResidency.Get(ctx))(&resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		bodyDR := apiclient.CreateWorkspaceRequest_DataResidency{
			AllowedInferenceGeos: &apiclient.CreateWorkspaceRequest_DataResidency_AllowedInferenceGeos{},
		}
		if !dr.AllowedInferenceGeos.IsNull() && !dr.AllowedInferenceGeos.IsUnknown() {
			_ = bodyDR.AllowedInferenceGeos.FromCreateWorkspaceRequestDataResidencyAllowedInferenceGeos0(fwdiag.Merge(dr.AllowedInferenceGeos.Get(ctx))(&resp.Diagnostics))
		} else if !dr.AllowedInferenceGeosUnrestricted.IsNull() && !dr.AllowedInferenceGeosUnrestricted.IsUnknown() {
			_ = bodyDR.AllowedInferenceGeos.FromCreateWorkspaceRequestDataResidencyAllowedInferenceGeos1(apiclient.CreateWorkspaceRequestDataResidencyAllowedInferenceGeos1Unrestricted)
		}
		if !dr.DefaultInferenceGeo.IsNull() && !dr.DefaultInferenceGeo.IsUnknown() {
			bodyDR.DefaultInferenceGeo.Set(apiclient.CreateWorkspaceRequestDataResidencyDefaultInferenceGeo(dr.DefaultInferenceGeo.ValueString()))
		} else {
			bodyDR.DefaultInferenceGeo.SetUnspecified()
		}
		if !dr.WorkspaceGeo.IsNull() && !dr.WorkspaceGeo.IsUnknown() {
			bodyDR.WorkspaceGeo.Set(apiclient.CreateWorkspaceRequestDataResidencyWorkspaceGeo(dr.WorkspaceGeo.ValueString()))
		} else {
			bodyDR.WorkspaceGeo.SetUnspecified()
		}

		body.DataResidency.Set(bodyDR)
	} else {
		body.DataResidency.SetUnspecified()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	workspace := fwdiag.Merge(apiclient.CreateJSON200(r.client.CreateWorkspaceWithResponse(
		ctx,
		nil,
		body,
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, *workspace)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkspaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspace := fwdiag.Merge(apiclient.ReadJSON200(r.client.GetWorkspaceWithResponse(
		ctx,
		data.Id.ValueString(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, *workspace)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkspaceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.UpdateWorkspaceJSONRequestBody{
		Name: data.Name.ValueStringPointer(),
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		body.Tags.Set(fwdiag.Merge(data.Tags.Get(ctx))(&resp.Diagnostics))
	} else {
		body.Tags.SetUnspecified()
	}

	if !data.DataResidency.IsNull() && !data.DataResidency.IsUnknown() {
		dr := fwdiag.Merge(data.DataResidency.Get(ctx))(&resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		var bodyDR apiclient.UpdateWorkspaceRequest_DataResidency
		if !dr.AllowedInferenceGeos.IsNull() && !dr.AllowedInferenceGeos.IsUnknown() {
			_ = bodyDR.AllowedInferenceGeos.FromUpdateWorkspaceRequestDataResidencyAllowedInferenceGeos0(fwdiag.Merge(dr.AllowedInferenceGeos.Get(ctx))(&resp.Diagnostics))
		} else if !dr.AllowedInferenceGeosUnrestricted.IsNull() && !dr.AllowedInferenceGeosUnrestricted.IsUnknown() {
			_ = bodyDR.AllowedInferenceGeos.FromUpdateWorkspaceRequestDataResidencyAllowedInferenceGeos1(apiclient.UpdateWorkspaceRequestDataResidencyAllowedInferenceGeos1Unrestricted)
		}
		if !dr.DefaultInferenceGeo.IsNull() && !dr.DefaultInferenceGeo.IsUnknown() {
			bodyDR.DefaultInferenceGeo.Set(apiclient.UpdateWorkspaceRequestDataResidencyDefaultInferenceGeo(dr.DefaultInferenceGeo.ValueString()))
		}

		body.DataResidency.Set(bodyDR)
	} else {
		body.DataResidency.SetUnspecified()
	}

	workspace := fwdiag.Merge(apiclient.UpdateJSON200(r.client.UpdateWorkspaceWithResponse(
		ctx,
		data.Id.ValueString(),
		body,
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, *workspace)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkspaceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = fwdiag.Merge(apiclient.DeleteJSON200(r.client.ArchiveWorkspaceWithResponse(
		ctx,
		data.Id.ValueString(),
	)))(&resp.Diagnostics)
}

func (r *WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
