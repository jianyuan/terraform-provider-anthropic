package provider

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwresource"
)

var _ resource.Resource = &ServiceAccountWorkspaceMemberResource{}
var _ resource.ResourceWithConfigure = &ServiceAccountWorkspaceMemberResource{}
var _ resource.ResourceWithImportState = &ServiceAccountWorkspaceMemberResource{}

func NewServiceAccountWorkspaceMemberResource() resource.Resource {
	return &ServiceAccountWorkspaceMemberResource{}
}

type ServiceAccountWorkspaceMemberResource struct {
	baseResource
}

func (r *ServiceAccountWorkspaceMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_workspace_member"
}

func (r *ServiceAccountWorkspaceMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = serviceAccountWorkspaceMemberSchema().GetResource(ctx)
}

func (r *ServiceAccountWorkspaceMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	fwresource.Create(ctx, func(data *ServiceAccountWorkspaceMemberModel, body *anthropic.OrganizationWorkspaceServiceAccountAddParams) (*anthropic.ServiceAccountWorkspaceMember, error) {
		return r.authTokenClient.Organization.Workspaces.ServiceAccounts.Add(ctx, data.WorkspaceId.ValueString(), *body)
	}, req, resp)
}

func (r *ServiceAccountWorkspaceMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	fwresource.Read(ctx, func(data *ServiceAccountWorkspaceMemberModel) (*anthropic.ServiceAccountWorkspaceMember, error) {
		return r.authTokenClient.Organization.Workspaces.ServiceAccounts.Get(
			ctx,
			data.ServiceAccountId.ValueString(),
			anthropic.OrganizationWorkspaceServiceAccountGetParams{
				WorkspaceID: data.WorkspaceId.ValueString(),
			},
		)
	}, req, resp)
}

func (r *ServiceAccountWorkspaceMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	fwresource.Update(ctx, func(data *ServiceAccountWorkspaceMemberModel, body *anthropic.OrganizationWorkspaceServiceAccountUpdateParams) (*anthropic.ServiceAccountWorkspaceMember, error) {
		return r.authTokenClient.Organization.Workspaces.ServiceAccounts.Update(ctx, data.ServiceAccountId.ValueString(), *body)
	}, req, resp)
}

func (r *ServiceAccountWorkspaceMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	fwresource.Delete(ctx, func(data *ServiceAccountWorkspaceMemberModel) error {
		_, err := r.authTokenClient.Organization.Workspaces.ServiceAccounts.Remove(
			ctx,
			data.ServiceAccountId.ValueString(),
			anthropic.OrganizationWorkspaceServiceAccountRemoveParams{
				WorkspaceID: data.WorkspaceId.ValueString(),
			},
		)
		return err
	}, req, resp)
}

func (r *ServiceAccountWorkspaceMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	workspaceId, userId, err := SplitTwoPartId(req.ID, "workspace_id", "service_account_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Error parsing ID: %s", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("workspace_id"), workspaceId,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("service_account_id"), userId,
	)...)
}
