package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
)

func TestAccUsersDataSource(t *testing.T) {
	rn := "data.anthropic_users.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("users"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":       knownvalue.StringExact(acctest.TestUserId),
							"email":    knownvalue.NotNull(),
							"name":     knownvalue.NotNull(),
							"role":     knownvalue.NotNull(),
							"added_at": knownvalue.NotNull(),
						}),
					})),
				},
			},
		},
	})
}

var testAccUsersDataSourceConfig = `
data "anthropic_users" "test" {
}
`
