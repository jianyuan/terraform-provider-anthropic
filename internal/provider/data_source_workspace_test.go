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

func TestAccWorkspaceDataSource(t *testing.T) {
	rn := "data.anthropic_workspace.test"
	workspaceName := sdkacctest.RandomWithPrefix("tf-workspace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig(workspaceName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(workspaceName)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("created_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("archived_at"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("display_color"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("compartment_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("data_residency"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"allowed_inference_geos":              knownvalue.Null(),
						"allowed_inference_geos_unrestricted": knownvalue.Bool(true),
						"default_inference_geo":               knownvalue.StringExact("global"),
						"workspace_geo":                       knownvalue.StringExact("us"),
					})),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("external_key_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tags"), knownvalue.MapExact(map[string]knownvalue.Check{
						"env":  knownvalue.StringExact("prod"),
						"team": knownvalue.StringExact("platform"),
					})),
				},
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "anthropic_workspace" "test" {
	name = %[1]q

	tags = {
		env = "prod"
		team = "platform"
	}
}

data "anthropic_workspace" "test" {
	id = anthropic_workspace.test.id
}
`, workspaceName)
}
