package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
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

	body := anthropic.OrganizationInviteNewParams{
		Email: data.Email.ValueString(),
		Role:  anthropic.OrganizationInviteNewParamsRole(data.Role.ValueString()),
	}
	if fwtypes.IsKnown(data.RbacGroupIds) {
		body.RBACGroupIDs = fwdiag.Merge(data.RbacGroupIds.Get(ctx))(&resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	invite, err := r.apiKeyClient.Organization.Invites.New(ctx, body)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewResourceCreateErrorDiagnostic(err))
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

	invite, err := r.apiKeyClient.Organization.Invites.Get(ctx, data.Id.ValueString())
	if err != nil {
		if apiclient.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.Append(fwdiag.NewResourceReadErrorDiagnostic(err))
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

	_, err := r.apiKeyClient.Organization.Invites.Delete(ctx, data.Id.ValueString())
	if err != nil {
		if !apiclient.IsNotFoundError(err) {
			resp.Diagnostics.Append(fwdiag.NewResourceDeleteErrorDiagnostic(err))
		}
	}
}

func (r *OrganizationInviteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
