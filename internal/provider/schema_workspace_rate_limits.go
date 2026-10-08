package provider

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func workspaceRateLimitsSchema() superschema.Schema {
	return superschema.Schema{
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "List a workspace's rate limits.\n\nBy default, returns only the groups and limiter types that have a workspace-level override. With `query.include_inherited = true`, returns every group with organization-level limits the workspace can see, listing for each the values it inherits from the organization as well as its own overrides. Each value's source says which it is.",
		},
		Attributes: superschema.Attributes{
			"workspace_id": superschema.StringAttribute{
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "The ID of the workspace.",
					Required:            true,
				},
			},
			"query": superschema.SuperSingleNestedAttributeOf[WorkspaceRateLimitsDataSourceModel_Query]{
				DataSource: &schemaD.SingleNestedAttribute{
					MarkdownDescription: "Query filters.",
					Optional:            true,
				},
				Attributes: superschema.Attributes{
					"group_type": superschema.StringAttribute{
						DataSource: &schemaD.StringAttribute{
							MarkdownDescription: "Filter by group type.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(anthropic.BetaOrganizationWorkspaceRateLimitListParamsGroupTypeBatch),
									string(anthropic.BetaOrganizationWorkspaceRateLimitListParamsGroupTypeFiles),
									string(anthropic.BetaOrganizationWorkspaceRateLimitListParamsGroupTypeModelGroup),
									string(anthropic.BetaOrganizationWorkspaceRateLimitListParamsGroupTypeSkills),
									string(anthropic.BetaOrganizationWorkspaceRateLimitListParamsGroupTypeTokenCount),
									string(anthropic.BetaOrganizationWorkspaceRateLimitListParamsGroupTypeWebSearch),
								),
							},
						},
					},
					"include_inherited": superschema.BoolAttribute{
						DataSource: &schemaD.BoolAttribute{
							MarkdownDescription: "Also list the limiter values the workspace inherits from the organization, including groups with no workspace-level override.",
							Optional:            true,
						},
					},
				},
			},
			"workspace_rate_limits": superschema.SuperListNestedAttributeOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit]{
				DataSource: &schemaD.ListNestedAttribute{
					MarkdownDescription: "The rate limits for the workspace.",
					Computed:            true,
				},
				Attributes: superschema.Attributes{
					"group": superschema.SuperSingleNestedAttributeOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Group]{
						DataSource: &schemaD.SingleNestedAttribute{
							MarkdownDescription: "The rate-limit group this entry's limits apply to.",
							Computed:            true,
						},
						Attributes: superschema.Attributes{
							"type": superschema.StringAttribute{
								DataSource: &schemaD.StringAttribute{
									MarkdownDescription: "The type of the group.",
									Computed:            true,
								},
							},
							"id": superschema.StringAttribute{
								DataSource: &schemaD.StringAttribute{
									MarkdownDescription: "The ID of the group.",
									Computed:            true,
								},
							},
							"display_name": superschema.StringAttribute{
								DataSource: &schemaD.StringAttribute{
									MarkdownDescription: "Human-readable name of the model group (for example, `Claude Sonnet 4.x`). For display only; it may change. Only available when `type = \"model_group\"`.",
									Computed:            true,
								},
							},
						},
					},
					"limits": superschema.SuperListNestedAttributeOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit]{
						DataSource: &schemaD.ListNestedAttribute{
							MarkdownDescription: "The workspace's limiter values for this group. By default only the limiter types with a workspace-level override are listed. With `query.include_inherited` set to `true`, the limiter types the workspace inherits from the organization are listed too, each marked by `source`.",
							Computed:            true,
						},
						Attributes: superschema.Attributes{
							"type": superschema.StringAttribute{
								DataSource: &schemaD.StringAttribute{
									MarkdownDescription: "The limiter type (for example, `requests_per_minute` or `input_tokens_per_minute`).",
									Computed:            true,
								},
							},
							"org_limit": superschema.Int64Attribute{
								DataSource: &schemaD.Int64Attribute{
									MarkdownDescription: "The organization-level value for the same limiter type, for reference. `null` when the organization has no limit configured for this limiter type.",
									Computed:            true,
								},
							},
							"source": superschema.SuperSingleNestedAttributeOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit_Source]{
								DataSource: &schemaD.SingleNestedAttribute{
									MarkdownDescription: "Where `value` comes from. `organization` values are listed only when `query.include_inherited` is `true`, and then `value` equals `org_limit`.",
									Computed:            true,
								},
								Attributes: superschema.Attributes{
									"type": superschema.StringAttribute{
										DataSource: &schemaD.StringAttribute{
											MarkdownDescription: "The type of the source.",
											Computed:            true,
										},
									},
								},
							},
							"value": superschema.Int64Attribute{
								DataSource: &schemaD.Int64Attribute{
									MarkdownDescription: "The workspace's value for this limiter type: the workspace-level override when `source.type` is `workspace`, otherwise the organization's value.",
									Computed:            true,
								},
							},
						},
					},
					"models": superschema.SuperListAttributeOf[string]{
						DataSource: &schemaD.ListAttribute{
							MarkdownDescription: "Model names this entry's limits apply to, including aliases. `null` when `group_type` is not `\"model_group\"`.",
							Computed:            true,
						},
					},
					"rate_limit_id": superschema.StringAttribute{
						DataSource: &schemaD.StringAttribute{
							MarkdownDescription: "The `id` of the organization's RateLimit entry this entry applies to.",
							Computed:            true,
						},
					},
					"workspace_id": superschema.StringAttribute{
						DataSource: &schemaD.StringAttribute{
							MarkdownDescription: "ID of the Workspace this entry applies to.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}
