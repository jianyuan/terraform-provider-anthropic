package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
	"github.com/jianyuan/terraform-provider-anthropic/internal/tfutils"
)

var _ resource.Resource = &OrganizationInviteResource{}
var _ resource.ResourceWithConfigure = &OrganizationInviteResource{}
var _ resource.ResourceWithImportState = &OrganizationInviteResource{}

func NewOrganizationInviteResource() resource.Resource {
	return &OrganizationInviteResource{}
}

type OrganizationInviteResource struct {
	baseResource
}

func (r *OrganizationInviteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_invite"
}

func (r *OrganizationInviteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = organizationInviteSchema().GetResource(ctx)
}

func (r *OrganizationInviteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationInviteModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.CreateInviteJSONRequestBody{
		Email: data.Email.ValueString(),
		Role:  apiclient.CreateInviteRequestRole(data.Role.ValueString()),
	}
	if fwtypes.IsKnown(data.RbacGroupIds) {
		body.RbacGroupIds = new(tfutils.MergeDiagnostics(data.RbacGroupIds.Get(ctx))(&resp.Diagnostics))
	}
	if resp.Diagnostics.HasError() {
		return
	}

	invite := fwdiag.Merge(apiclient.CreateJSON200(r.client.CreateInviteWithResponse(
		ctx,
		body,
		r.WithApiKeyRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *invite)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationInviteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationInviteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	invite := fwdiag.Merge(apiclient.ReadJSON200(r.client.GetInviteWithResponse(
		ctx,
		data.Id.ValueString(),
		r.WithApiKeyRequestEditorFn(),
	)))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		if resp.Diagnostics.Contains(fwdiag.ErrorDiagnosticNotFound) {
			resp.State.RemoveResource(ctx)
		}
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *invite)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationInviteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Updates are not supported for invites - any change requires replacement
	resp.Diagnostics.AddError("Update Not Supported", "Organization invites cannot be updated. Any changes require creating a new invite.")
}

func (r *OrganizationInviteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationInviteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_ = fwdiag.Merge(apiclient.DeleteJSON200(r.client.DeleteInviteWithResponse(
		ctx,
		data.Id.ValueString(),
		r.WithApiKeyRequestEditorFn(),
	)))(&resp.Diagnostics)
}

func (r *OrganizationInviteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
