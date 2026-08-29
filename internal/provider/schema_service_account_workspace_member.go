package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func serviceAccountWorkspaceMemberSchema() superschema.Schema {
	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Add a service account to a workspace with the given `workspace_role`.\n\nRequires an OAuth bearer. Admin API keys are not accepted.",
		},
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "Retrieve a service account's membership in a workspace.",
		},
		Attributes: superschema.Attributes{
			"workspace_id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "ID of the workspace.",
					Required:            true,
				},
				Resource: &schemaR.StringAttribute{
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"service_account_id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "ID of the service account.",
					Required:            true,
				},
				Resource: &schemaR.StringAttribute{
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"workspace_role": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Role of the service account in the workspace.",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf("workspace_admin", "workspace_developer", "workspace_restricted_developer", "workspace_user"),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"created_by_actor_id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Tagged ID (`user_...`/`svac_...`) of the actor who created this membership.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
			"implicit": superschema.BoolAttribute{
				Common: &schemaR.BoolAttribute{
					MarkdownDescription: "True when this is the implicit default-workspace membership every service account has when no explicit membership exists. Implicit memberships have role `workspace_user` and cannot be removed.",
					Computed:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}
