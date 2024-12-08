package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/go-utils/ptr"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type WorkspaceMembersDataSourceModel struct {
	Id      types.String           `tfsdk:"id"`
	Members []WorkspaceMemberModel `tfsdk:"members"`
}

func (m *WorkspaceMembersDataSourceModel) Fill(members []apiclient.WorkspaceMember) error {
	m.Members = make([]WorkspaceMemberModel, len(members))
	for i, u := range members {
		if err := m.Members[i].Fill(u); err != nil {
			return err
		}
	}

	return nil
}

func NewWorkspaceMembersDataSource() datasource.DataSource {
	return &WorkspaceMembersDataSource{}
}

var _ datasource.DataSource = &WorkspaceMembersDataSource{}

type WorkspaceMembersDataSource struct {
	baseDataSource
}

func (d *WorkspaceMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_members"
}

func (d *WorkspaceMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all members of the workspace.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Workspace.",
				Required:            true,
			},
			"members": schema.SetNestedAttribute{
				MarkdownDescription: "List of members.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"workspace_id": schema.StringAttribute{
							MarkdownDescription: "ID of the Workspace to which the member belongs.",
							Computed:            true,
						},
						"user_id": schema.StringAttribute{
							MarkdownDescription: "ID of the user who is a member of the Workspace.",
							Computed:            true,
						},
						"workspace_role": schema.StringAttribute{
							MarkdownDescription: "Role of the new Workspace Member. Must be one of `workspace_user`, `workspace_developer`, or `workspace_admin`.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *WorkspaceMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkspaceMembersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var members []apiclient.WorkspaceMember
	params := &apiclient.ListWorkspaceMembersParams{
		Limit: ptr.Ptr(100),
	}

	for {
		httpResp, err := d.client.ListWorkspaceMembersWithResponse(
			ctx,
			data.Id.ValueString(),
			params,
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

		members = append(members, httpResp.JSON200.Data...)

		if !httpResp.JSON200.HasMore || httpResp.JSON200.LastId == nil {
			break
		}

		params.AfterId = httpResp.JSON200.LastId
	}

	if err := data.Fill(members); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill data, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
