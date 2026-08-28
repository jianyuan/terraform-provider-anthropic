package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func workspaceSchema() superschema.Schema {
	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Manage a Workspace.",
		},
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "Get information about a Workspace.",
		},
		Attributes: map[string]superschema.Attribute{
			"id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "ID of the Workspace.",
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
					MarkdownDescription: "Name of the Workspace.",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"created_at": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "RFC 3339 datetime string indicating when the Workspace was created.",
					Computed:            true,
				},
			},
			"archived_at": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "RFC 3339 datetime string indicating when the Workspace was archived, or null if the Workspace is not archived.",
					Computed:            true,
				},
			},
			"display_color": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Hex color code representing the Workspace in the Anthropic Console.",
					Computed:            true,
				},
			},
			"compartment_id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Identifier for this Workspace's encryption compartment. When you configure a customer-managed encryption key (CMEK) on AWS, reference this value in your KMS key-policy condition so the key is scoped to this compartment. On GCP and Azure, Anthropic enforces the compartment binding automatically; you do not need to reference this value in your key configuration. See the CMEK integration guide for the required key configuration, including the value used during key validation.",
					Computed:            true,
				},
			},
			"data_residency": superschema.SuperSingleNestedAttributeOf[WorkspaceModelDataResidency]{
				Common: &schemaR.SingleNestedAttribute{
					MarkdownDescription: "Data residency configuration.",
				},
				Resource: &schemaR.SingleNestedAttribute{
					Optional: true,
					Computed: true,
				},
				DataSource: &schemaD.SingleNestedAttribute{
					Computed: true,
				},
				Attributes: map[string]superschema.Attribute{
					"allowed_inference_geos": superschema.SuperSetAttributeOf[string]{
						Common: &schemaR.SetAttribute{
							MarkdownDescription: "Permitted inference geo values.",
						},
						Resource: &schemaR.SetAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Set{
								setvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("allowed_inference_geos_unrestricted")),
							},
						},
						DataSource: &schemaD.SetAttribute{
							Computed: true,
						},
					},
					"allowed_inference_geos_unrestricted": superschema.BoolAttribute{
						Common: &schemaR.BoolAttribute{
							MarkdownDescription: "All geos available for inference.",
						},
						Resource: &schemaR.BoolAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Bool{
								boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("allowed_inference_geos")),
							},
						},
						DataSource: &schemaD.BoolAttribute{
							Computed: true,
						},
					},
					"default_inference_geo": superschema.StringAttribute{
						Common: &schemaR.StringAttribute{
							MarkdownDescription: "Default inference geo applied when requests omit the parameter.",
						},
						Resource: &schemaR.StringAttribute{
							Optional: true,
							Computed: true,
						},
						DataSource: &schemaD.StringAttribute{
							Computed: true,
						},
					},
					"workspace_geo": superschema.StringAttribute{
						Common: &schemaR.StringAttribute{
							MarkdownDescription: "Geographic region for workspace data storage. Immutable after creation.",
						},
						Resource: &schemaR.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						DataSource: &schemaD.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"external_key_id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "ID of the customer-managed encryption key (CMEK) configuration to use for this Workspace. Setting this field requires CMEK to be enabled for your organization. When set, data stored for this Workspace is encrypted with the referenced key. Create key configurations with the External Keys API. This field is write-once: once a key is attached to a Workspace it cannot be detached or replaced. To rotate key material, rotate the underlying key on your cloud KMS; the `external_key_id` stays the same.",
				},
				Resource: &schemaR.StringAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"tags": superschema.SuperMapAttributeOf[string]{
				Common: &schemaR.MapAttribute{
					MarkdownDescription: "User-defined tags as string key-value pairs. Keys may not begin with `anthropic`.",
				},
				Resource: &schemaR.MapAttribute{
					Optional: true,
					Computed: true,
				},
				DataSource: &schemaD.MapAttribute{
					Computed: true,
				},
			},
		},
	}
}
