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

func TestAccWorkspaceDataSource(t *testing.T) {
	rn := "data.anthropic_workspace.test"
	workspaceName := acctest.RandomWithPrefix("tf-workspace")

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
				},
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "anthropic_workspace" "test" {
	name = %[1]q
}

data "anthropic_workspace" "test" {
	id = anthropic_workspace.test.id
}
`, workspaceName)
}
