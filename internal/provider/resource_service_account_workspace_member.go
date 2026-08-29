package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
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
	var data ServiceAccountWorkspaceMemberModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := fwdiag.Merge(data.ToAPIForCreate(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	sawm := fwdiag.Merge(apiclient.CreateJSON200(r.client.CreateServiceAccountWorkspaceMemberWithResponse(
		ctx,
		data.WorkspaceId.ValueString(),
		nil,
		*body,
		r.WithAuthTokenRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromCreateAPI(ctx, *sawm)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountWorkspaceMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceAccountWorkspaceMemberModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sawm := fwdiag.Merge(apiclient.ReadJSON200(r.client.GetServiceAccountWorkspaceMemberWithResponse(
		ctx,
		data.WorkspaceId.ValueString(),
		data.ServiceAccountId.ValueString(),
		nil,
		r.WithAuthTokenRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromReadAPI(ctx, *sawm)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountWorkspaceMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServiceAccountWorkspaceMemberModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := fwdiag.Merge(data.ToAPIForUpdate(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	sawm := fwdiag.Merge(apiclient.UpdateJSON200(r.client.UpdateServiceAccountWorkspaceMemberWithResponse(
		ctx,
		data.WorkspaceId.ValueString(),
		data.ServiceAccountId.ValueString(),
		nil,
		*body,
		r.WithAuthTokenRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromUpdateAPI(ctx, *sawm)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountWorkspaceMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceAccountWorkspaceMemberModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = fwdiag.Merge(apiclient.DeleteJSON200(r.client.DeleteServiceAccountWorkspaceMemberWithResponse(
		ctx,
		data.WorkspaceId.ValueString(),
		data.ServiceAccountId.ValueString(),
		nil,
		r.WithAuthTokenRequestEditorFn(),
	)))(&resp.Diagnostics)
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
