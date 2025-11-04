package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

func NewOrganizationInviteResource() resource.Resource {
	return &OrganizationInviteResource{}
}

var _ resource.Resource = &OrganizationInviteResource{}
var _ resource.ResourceWithImportState = &OrganizationInviteResource{}

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
				MarkdownDescription: "Unique identifier for the invite.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address of the person being invited.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role to assign to the invited user. Must be one of `user`, `developer`, `billing`, `admin`, or `claude_code_user`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("user", "developer", "billing", "admin", "claude_code_user"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current status of the invite (e.g., pending, accepted, expired).",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the invite was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the invite expires.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	httpResp, err := r.client.CreateInviteWithResponse(
		ctx,
		apiclient.CreateInviteJSONRequestBody{
			Email: data.Email.ValueString(),
			Role:  data.Role.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create invite, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create invite, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create invite, got empty response body")
		return
	}

	if err := data.Fill(*httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
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

	httpResp, err := r.client.GetInviteWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read invite, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read invite, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read invite, got empty response body")
		return
	}

	if err := data.Fill(*httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
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

	httpResp, err := r.client.DeleteInviteWithResponse(
		ctx,
		data.Id.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete invite, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete invite, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
}

func (r *OrganizationInviteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
