package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func organizationInvitesSchema() superschema.Schema {
	inviteAttributes := organizationInviteSchema().Attributes

	return superschema.Schema{
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "List all pending invites in the Organization.",
		},
		Attributes: superschema.Attributes{
			"invites": superschema.SuperSetNestedAttributeOf[OrganizationInviteModel]{
				DataSource: &schemaD.SetNestedAttribute{
					MarkdownDescription: "List of pending organization invites.",
					Computed:            true,
				},
				Attributes: inviteAttributes,
			},
		},
	}
}
