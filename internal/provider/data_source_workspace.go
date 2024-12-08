package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func NewWorkspaceDataSource() datasource.DataSource {
	return &WorkspaceDataSource{}
}

var _ datasource.DataSource = &WorkspaceDataSource{}

type WorkspaceDataSource struct {
	baseDataSource
}

func (d *WorkspaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *WorkspaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about a Workspace.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Workspace.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Workspace.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Workspace was created.",
				Computed:            true,
			},
			"archived_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Workspace was archived, or null if the Workspace is not archived.",
				Computed:            true,
			},
			"display_color": schema.StringAttribute{
				MarkdownDescription: "Hex color code representing the Workspace in the Anthropic Console.",
				Computed:            true,
			},
		},
	}
}

func (d *WorkspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkspaceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetWorkspaceWithResponse(
		ctx,
		data.Id.ValueString(),
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
