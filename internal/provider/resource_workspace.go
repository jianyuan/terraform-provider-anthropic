package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwresource"
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
	fwresource.Create(ctx, func(_ *WorkspaceModel, body *anthropic.OrganizationWorkspaceNewParams) (*anthropic.Workspace, error) {
		return r.apiKeyClient.Organization.Workspaces.New(ctx, *body)
	}, req, resp)
}

func (r *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	fwresource.Read(ctx, func(data *WorkspaceModel) (*anthropic.Workspace, error) {
		return r.apiKeyClient.Organization.Workspaces.Get(ctx, data.Id.ValueString())
	}, req, resp)
}

func (r *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	fwresource.Update(ctx, func(data *WorkspaceModel, body *anthropic.OrganizationWorkspaceUpdateParams) (*anthropic.Workspace, error) {
		return r.apiKeyClient.Organization.Workspaces.Update(ctx, data.Id.ValueString(), *body)
	}, req, resp)
}

func (r *WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	fwresource.Delete(ctx, func(data *WorkspaceModel) error {
		_, err := r.apiKeyClient.Organization.Workspaces.Archive(ctx, data.Id.ValueString())
		return err
	}, req, resp)
}

func (r *WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
