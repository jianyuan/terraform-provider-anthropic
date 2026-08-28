package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
)

func TestAccWorkspacesDataSource(t *testing.T) {
	rn := "data.anthropic_workspaces.test"
	workspaceName := sdkacctest.RandomWithPrefix("tf-workspace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDataSourceConfig(workspaceName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("workspaces"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":             knownvalue.NotNull(),
							"name":           knownvalue.StringExact(workspaceName),
							"created_at":     knownvalue.NotNull(),
							"archived_at":    knownvalue.Null(),
							"display_color":  knownvalue.NotNull(),
							"compartment_id": knownvalue.NotNull(),
							"data_residency": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"allowed_inference_geos":              knownvalue.Null(),
								"allowed_inference_geos_unrestricted": knownvalue.Bool(true),
								"default_inference_geo":               knownvalue.StringExact("global"),
								"workspace_geo":                       knownvalue.StringExact("us"),
							}),
							"external_key_id": knownvalue.Null(),
							"tags": knownvalue.MapExact(map[string]knownvalue.Check{
								"env":  knownvalue.StringExact("prod"),
								"team": knownvalue.StringExact("platform"),
							}),
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

	tags = {
		env = "prod"
		team = "platform"
	}
}

data "anthropic_workspaces" "test" {
	depends_on = [anthropic_workspace.test]
}
`, workspaceName)
}
