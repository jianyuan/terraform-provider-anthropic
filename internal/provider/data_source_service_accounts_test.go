package provider

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
)

func TestAccServiceAccountsDataSource(t *testing.T) {
	rn := "data.anthropic_service_accounts.test"
	name := sdkacctest.RandomWithPrefix("tf-sa")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountsDataSourceConfig(name, "admin"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("service_accounts"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"name":              knownvalue.StringExact(name),
							"organization_role": knownvalue.StringExact("admin"),
						}),
					})),
				},
			},
		},
	})
}

func testAccServiceAccountsDataSourceConfig(name, orgRole string) string {
	return testAccServiceAccountResourceConfig(name, orgRole, "") + `
data "anthropic_service_accounts" "test" {
	depends_on = [anthropic_service_account.test]
}
`
}
