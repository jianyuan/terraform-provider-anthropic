package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func workspacesSchema() superschema.Schema {
	workspaceAttributes := workspaceSchema().Attributes
	workspaceAttributes["id"].(superschema.StringAttribute).DataSource.Required = false
	workspaceAttributes["id"].(superschema.StringAttribute).DataSource.Computed = true

	return superschema.Schema{
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "List all workspaces in the organization.",
		},
		Attributes: map[string]superschema.Attribute{
			"workspaces": superschema.SuperSetNestedAttributeOf[WorkspaceModel]{
				DataSource: &schemaD.SetNestedAttribute{
					MarkdownDescription: "List of workspaces.",
					Computed:            true,
				},
				Attributes: workspaceAttributes,
			},
		},
	}
}
