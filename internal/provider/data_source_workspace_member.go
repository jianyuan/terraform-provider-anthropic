package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func NewWorkspaceMemberDataSource() datasource.DataSource {
	return &WorkspaceMemberDataSource{}
}

var _ datasource.DataSource = &WorkspaceMemberDataSource{}

type WorkspaceMemberDataSource struct {
	baseDataSource
}

func (d *WorkspaceMemberDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_member"
}

func (d *WorkspaceMemberDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get a member of a Workspace.",

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
				Computed:            true,
			},
		},
	}
}

func (d *WorkspaceMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkspaceMemberModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetWorkspaceMemberWithResponse(
		ctx,
		data.WorkspaceId.ValueString(),
		data.UserId.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got status code: %d", httpResp.StatusCode()))
		return
	}

	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read, got empty response body")
		return
	}

	if err := data.Fill(*httpResp.JSON200); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
