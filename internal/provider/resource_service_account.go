package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

var _ resource.Resource = &ServiceAccountResource{}
var _ resource.ResourceWithConfigure = &ServiceAccountResource{}
var _ resource.ResourceWithImportState = &ServiceAccountResource{}

func NewServiceAccountResource() resource.Resource {
	return &ServiceAccountResource{}
}

type ServiceAccountResource struct {
	baseResource
}

func (r *ServiceAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *ServiceAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = serviceAccountSchema().GetResource(ctx)
}

func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceAccountModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := fwdiag.Merge(data.ToAPIForCreate(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	sa := fwdiag.Merge(apiclient.CreateJSON200(r.client.CreateServiceAccountWithResponse(
		ctx,
		nil,
		*body,
		r.WithAuthTokenRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *sa)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceAccountModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sa := fwdiag.Merge(apiclient.ReadJSON200(r.client.GetServiceAccountWithResponse(
		ctx,
		data.Id.ValueString(),
		nil,
		r.WithAuthTokenRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *sa)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServiceAccountModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := fwdiag.Merge(data.ToAPIForUpdate(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	sa := fwdiag.Merge(apiclient.UpdateJSON200(r.client.UpdateServiceAccountWithResponse(
		ctx,
		data.Id.ValueString(),
		nil,
		*body,
		r.WithAuthTokenRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *sa)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceAccountModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = fwdiag.Merge(apiclient.DeleteJSON200(r.client.ArchiveServiceAccountWithResponse(
		ctx,
		data.Id.ValueString(),
		nil,
		r.WithAuthTokenRequestEditorFn(),
	)))(&resp.Diagnostics)
}

func (r *ServiceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
