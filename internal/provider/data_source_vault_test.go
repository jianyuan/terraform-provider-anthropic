package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
)

func TestAccVaultDataSource(t *testing.T) {
	rn := "data.anthropic_vault.test"
	displayName := acctest.RandomWithPrefix("tf-vault")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultDataSourceConfig(displayName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("display_name"), knownvalue.StringExact(displayName)),
				},
			},
		},
	})
}

func testAccVaultDataSourceConfig(displayName string) string {
	return fmt.Sprintf(`
resource "anthropic_vault" "test" {
	display_name = %[1]q
}

data "anthropic_vault" "test" {
	id = anthropic_vault.test.id
}
`, displayName)
}
