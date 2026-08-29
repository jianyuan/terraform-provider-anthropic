package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func workspaceMembersSchema() superschema.Schema {
	memberAttributes := workspaceMemberSchema().Attributes
	//nolint:forcetypeassert
	memberAttributes["workspace_id"].(superschema.StringAttribute).Common.Required = false
	//nolint:forcetypeassert
	memberAttributes["workspace_id"].(superschema.StringAttribute).Common.Computed = true
	//nolint:forcetypeassert
	memberAttributes["user_id"].(superschema.StringAttribute).Common.Required = false
	//nolint:forcetypeassert
	memberAttributes["user_id"].(superschema.StringAttribute).Common.Computed = true

	return superschema.Schema{
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "List all members of the workspace.",
		},
		Attributes: superschema.Attributes{
			"id": superschema.StringAttribute{
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "ID of the Workspace.",
					Required:            true,
				},
			},
			"members": superschema.SuperSetNestedAttributeOf[WorkspaceMemberModel]{
				DataSource: &schemaD.SetNestedAttribute{
					MarkdownDescription: "List of Workspace Members.",
					Computed:            true,
				},
				Attributes: memberAttributes,
			},
		},
	}
}
