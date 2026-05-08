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

func TestAccEnvironmentDataSource(t *testing.T) {
	rn := "data.anthropic_environment.test"
	name := acctest.RandomWithPrefix("tf-environment")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccEnvironmentDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "anthropic_environment" "test" {
	name = %[1]q
	config = {
		networking = {
			type = "unrestricted"
		}
	}
}

data "anthropic_environment" "test" {
	id = anthropic_environment.test.id
}
`, name)
}
