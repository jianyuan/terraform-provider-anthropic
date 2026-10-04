package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/avast/retry-go/v5"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwresource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
)

var _ resource.Resource = &WorkspaceMemberResource{}
var _ resource.ResourceWithConfigure = &WorkspaceMemberResource{}
var _ resource.ResourceWithImportState = &WorkspaceMemberResource{}

func NewWorkspaceMemberResource() resource.Resource {
	return &WorkspaceMemberResource{}
}

type WorkspaceMemberResource struct {
	baseResource
}

func (r *WorkspaceMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_member"
}

func (r *WorkspaceMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = workspaceMemberSchema().GetResource(ctx)
}

func (r *WorkspaceMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	fwresource.Create(ctx, func(data *WorkspaceMemberModel, body *anthropic.OrganizationWorkspaceMemberAddParams) (*anthropic.WorkspaceMember, error) {
		return r.apiKeyClient.Organization.Workspaces.Members.Add(ctx, data.WorkspaceId.ValueString(), *body)
	}, req, resp)
}

func (r *WorkspaceMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	fwresource.Read(ctx, func(data *WorkspaceMemberModel) (*anthropic.WorkspaceMember, error) {
		var item *anthropic.WorkspaceMember
		err := retry.New(
			retry.Context(ctx),
			retry.Attempts(10),
			retry.Delay(3*time.Second),
		).Do(func() error {
			var err error
			item, err = r.apiKeyClient.Organization.Workspaces.Members.Get(
				ctx,
				data.UserId.ValueString(),
				anthropic.OrganizationWorkspaceMemberGetParams{
					WorkspaceID: data.WorkspaceId.ValueString(),
				},
			)
			if err != nil {
				return err
			} else if fwtypes.IsKnown(data.WorkspaceRole) && (!item.JSON.WorkspaceRole.Valid() || string(item.WorkspaceRole) != data.WorkspaceRole.ValueString()) {
				return fmt.Errorf("unexpected workspace role: %s, expected: %s", item.WorkspaceRole, data.WorkspaceRole.ValueString())
			}
			return nil
		})
		return item, err
	}, req, resp)
}

func (r *WorkspaceMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	fwresource.Update(ctx, func(data *WorkspaceMemberModel, body *anthropic.OrganizationWorkspaceMemberUpdateParams) (*anthropic.WorkspaceMember, error) {
		return r.apiKeyClient.Organization.Workspaces.Members.Update(ctx, data.UserId.ValueString(), *body)
	}, req, resp)
}

func (r *WorkspaceMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	fwresource.Delete(ctx, func(data *WorkspaceMemberModel) error {
		_, err := r.apiKeyClient.Organization.Workspaces.Members.Remove(
			ctx,
			data.UserId.ValueString(),
			anthropic.OrganizationWorkspaceMemberRemoveParams{
				WorkspaceID: data.WorkspaceId.ValueString(),
			},
		)
		return err
	}, req, resp)
}

func (r *WorkspaceMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	workspaceId, userId, err := SplitTwoPartId(req.ID, "workspace_id", "user_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Error parsing ID: %s", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("workspace_id"), workspaceId,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("user_id"), userId,
	)...)
}
