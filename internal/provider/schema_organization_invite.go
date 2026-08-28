package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func organizationInviteSchema() superschema.Schema {
	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Organization invite resource. Manages invitations to join an organization.",
		},
		Attributes: superschema.Attributes{
			"id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "ID of the Invite.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
			"email": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Email of the User being invited.",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"role": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Organization role of the User. Must be one of `admin`, `billing`, `claude_code_user`, `developer`, `managed`, `membership_admin`, `owner`, `primary_owner`, `user`.",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"rbac_group_ids": superschema.SuperSetAttributeOf[string]{
				Common: &schemaR.SetAttribute{
					MarkdownDescription: "RBAC group IDs to assign to the User when the Invite is accepted. A non-empty array is accepted only for a Claude Enterprise organization with RBAC groups (beta), and requires the key to carry the `write:rbac_groups` scope.",
				},
				Resource: &schemaR.SetAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.Set{
						setplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.SetAttribute{
					Computed: true,
				},
			},
			"status": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Status of the Invite (e.g. `accepted`, `deleted`, `expired`, or `pending`).",
					Computed:            true,
				},
			},
			"accepted_at": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "RFC 3339 datetime string indicating when the Invite was accepted, or null.",
					Computed:            true,
				},
			},
			"expires_at": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "RFC 3339 datetime string indicating when the Invite expires.",
					Computed:            true,
				},
			},
			"invited_at": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "RFC 3339 datetime string indicating when the Invite was created.",
					Computed:            true,
				},
			},
		},
	}
}
