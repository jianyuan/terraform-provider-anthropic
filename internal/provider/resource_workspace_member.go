package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	retry "github.com/avast/retry-go/v5"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

func NewWorkspaceMemberResource() resource.Resource {
	return &WorkspaceMemberResource{}
}

var _ resource.Resource = &WorkspaceMemberResource{}
var _ resource.ResourceWithImportState = &WorkspaceMemberResource{}

type WorkspaceMemberResource struct {
	baseResource
}

func (r *WorkspaceMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_member"
}

func (r *WorkspaceMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Workspace member resource.",

		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Workspace to which the member belongs.",
				Required:            true,
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "ID of the user who is a member of the Workspace.",
				Required:            true,
			},
			"workspace_role": schema.StringAttribute{
				MarkdownDescription: "Role of the new Workspace Member. Must be one of `workspace_user`, `workspace_developer`, or `workspace_admin`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("workspace_user", "workspace_developer", "workspace_admin"),
				},
			},
		},
	}
}

func (r *WorkspaceMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkspaceMemberModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreateWorkspaceMemberWithResponse(
		ctx,
		data.WorkspaceId.ValueString(),
		apiclient.CreateWorkspaceMemberJSONRequestBody{
			UserId:        data.UserId.ValueString(),
			WorkspaceRole: data.WorkspaceRole.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create, got empty response body")
		return
	}

	if err := data.Fill(*httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkspaceMemberModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var httpResp *apiclient.GetWorkspaceMemberResponse
	err := retry.New(
		retry.Context(ctx),
		retry.Attempts(10),
		retry.Delay(3*time.Second),
	).Do(func() error {
		var err error
		httpResp, err = r.client.GetWorkspaceMemberWithResponse(
			ctx,
			data.WorkspaceId.ValueString(),
			data.UserId.ValueString(),
		)
		if err != nil {
			return err
		} else if httpResp.StatusCode() != http.StatusOK {
			return fmt.Errorf("status code %d: %s", httpResp.StatusCode(), string(httpResp.Body))
		} else if httpResp.JSON200 == nil {
			return fmt.Errorf("empty response body")
		} else if !data.WorkspaceRole.IsNull() && !data.WorkspaceRole.IsUnknown() && httpResp.JSON200.WorkspaceRole != data.WorkspaceRole.ValueString() {
			return fmt.Errorf("unexpected workspace role: %s, expected: %s", httpResp.JSON200.WorkspaceRole, data.WorkspaceRole.ValueString())
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
		return
	} else if httpResp == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read, got empty response body")
		return
	}

	if err := data.Fill(*httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkspaceMemberModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.UpdateWorkspaceMemberWithResponse(
		ctx,
		data.WorkspaceId.ValueString(),
		data.UserId.ValueString(),
		apiclient.UpdateWorkspaceMemberJSONRequestBody{
			WorkspaceRole: data.WorkspaceRole.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update, got empty response body")
		return
	}

	if err := data.Fill(*httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspaceMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkspaceMemberModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DeleteWorkspaceMemberWithResponse(
		ctx,
		data.WorkspaceId.ValueString(),
		data.UserId.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
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
