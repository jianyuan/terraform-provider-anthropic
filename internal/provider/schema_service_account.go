package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func serviceAccountSchema() superschema.Schema {
	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Manage a Service Account.",
		},
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "Retrieve a Service Account.",
		},
		Attributes: superschema.Attributes{
			"id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "ID of the service account.",
				},
				Resource: &schemaR.StringAttribute{
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Required: true,
				},
			},
			"name": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Slug identifier (lowercase, digits, hyphens). Unique within the organization.",
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
			"description": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Optional free-text description.",
				},
				Resource: &schemaR.StringAttribute{
					Optional: true,
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"organization_role": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Org-level role. Defaults to `developer`.",
				},
				Resource: &schemaR.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.OneOf("admin", "developer"),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"archived_at": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "If set, this service account is archived.",
					Computed:            true,
				},
			},
			"archived_by_actor_id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Tagged ID (`user_`/`svac_`) of the actor that archived this service account.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			}, "created_at": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "When this service account was created.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
			"created_by_actor_id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Tagged ID (`user_`/`svac_`) of the actor that created this service account.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
			"updated_at": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "When this service account was last updated.",
					Computed:            true,
				},
			},
			"updated_by_actor_id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Tagged ID (`user_`/`svac_`) of the actor that last updated this service account.",
					Computed:            true,
				},
			},
		},
	}
}
