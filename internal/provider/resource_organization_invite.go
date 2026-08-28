package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/tfutils"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
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
	resp.Schema = schema.Schema{
		MarkdownDescription: "Organization invite resource. Manages invitations to join an organization.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Invite.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email of the User being invited.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Organization role of the User. Must be one of `admin`, `billing`, `claude_code_user`, `developer`, `managed`, `membership_admin`, `owner`, `primary_owner`, `user`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("admin", "billing", "claude_code_user", "developer", "managed", "membership_admin", "owner", "primary_owner", "user"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rbac_group_ids": schema.SetAttribute{
				MarkdownDescription: "RBAC group IDs to assign to the User when the Invite is accepted. A non-empty array is accepted only for a Claude Enterprise organization with RBAC groups (beta), and requires the key to carry the `write:rbac_groups` scope.",
				Optional:            true,
				CustomType:          supertypes.NewSetTypeOf[string](ctx),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the Invite (e.g. `accepted`, `deleted`, `expired`, or `pending`).",
				Computed:            true,
			},
			"accepted_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Invite was accepted, or null.",
				Computed:            true,
			},
			"invited_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Invite was created.",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Invite expires.",
				Computed:            true,
			},
		},
	}
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
	if !data.RbacGroupIds.IsNull() && !data.RbacGroupIds.IsUnknown() {
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
