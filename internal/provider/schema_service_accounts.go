package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func serviceAccountsSchema() superschema.Schema {
	serviceAccountAttributes := serviceAccountSchema().Attributes
	//nolint:forcetypeassert
	serviceAccountAttributes["id"].(superschema.StringAttribute).DataSource.Required = false
	//nolint:forcetypeassert
	serviceAccountAttributes["id"].(superschema.StringAttribute).DataSource.Computed = true

	return superschema.Schema{
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "List all service accounts in the organization.",
		},
		Attributes: superschema.Attributes{
			"service_accounts": superschema.SuperSetNestedAttributeOf[ServiceAccountModel]{
				DataSource: &schemaD.SetNestedAttribute{
					MarkdownDescription: "List of service accounts.",
					Computed:            true,
				},
				Attributes: serviceAccountAttributes,
			},
		},
	}
}
