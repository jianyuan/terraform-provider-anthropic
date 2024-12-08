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

func TestAccWorkspacesDataSource(t *testing.T) {
	rn := "data.anthropic_workspaces.test"
	workspaceName := acctest.RandomWithPrefix("tf-workspace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfig(workspaceName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("workspaces"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":            knownvalue.NotNull(),
							"name":          knownvalue.StringExact(workspaceName),
							"created_at":    knownvalue.NotNull(),
							"archived_at":   knownvalue.Null(),
							"display_color": knownvalue.NotNull(),
						}),
					})),
				},
			},
		},
	})
}

func testAccWorkspacesDataSourceConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "anthropic_workspace" "test" {
	name = %[1]q
}

data "anthropic_workspaces" "test" {
	depends_on = [anthropic_workspace.test]
}
`, workspaceName)
}
