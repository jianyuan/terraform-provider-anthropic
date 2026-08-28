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

	body := fwdiag.Merge(data.ToAPIForCreate(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	workspace := fwdiag.Merge(apiclient.CreateJSON200(r.client.CreateWorkspaceWithResponse(
		ctx,
		nil,
		*body,
		r.WithApiKeyRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *workspace)...)
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
		r.WithApiKeyRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *workspace)...)
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

	body := fwdiag.Merge(data.ToAPIForUpdate(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	workspace := fwdiag.Merge(apiclient.UpdateJSON200(r.client.UpdateWorkspaceWithResponse(
		ctx,
		data.Id.ValueString(),
		*body,
		r.WithApiKeyRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *workspace)...)
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
		r.WithApiKeyRequestEditorFn(),
	)))(&resp.Diagnostics)
}

func (r *WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
