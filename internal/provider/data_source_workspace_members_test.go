package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
)

func TestAccWorkspaceMembersDataSource(t *testing.T) {
	rn := "data.anthropic_workspace_members.test"
	workspaceName := acctest.RandomWithPrefix("tf-workspace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceMembersDataSourceConfig(workspaceName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(rn, tfjsonpath.New("id"), "anthropic_workspace.test", tfjsonpath.New("id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("members"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"workspace_id":   knownvalue.NotNull(),
							"user_id":        knownvalue.StringExact(acctest.TestUserId),
							"workspace_role": knownvalue.StringExact("workspace_developer"),
						}),
					})),
				},
			},
		},
	})
}

func testAccWorkspaceMembersDataSourceConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "anthropic_workspace" "test" {
	name = %[1]q
}

data "anthropic_user" "test" {
	id = %[2]q
}

resource "anthropic_workspace_member" "test" {
	workspace_id   = anthropic_workspace.test.id
	user_id        = data.anthropic_user.test.id
	workspace_role = "workspace_developer"
}

data "anthropic_workspace_members" "test" {
	depends_on  = [anthropic_workspace_member.test]

	id = anthropic_workspace_member.test.workspace_id
}
`, workspaceName, acctest.TestUserId)
}
