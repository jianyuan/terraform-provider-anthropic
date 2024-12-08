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

func TestAccWorkspaceMemberDataSource(t *testing.T) {
	rn := "data.anthropic_workspace_member.test"
	workspaceName := acctest.RandomWithPrefix("tf-workspace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceMemberDataSourceConfig(workspaceName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(rn, tfjsonpath.New("workspace_id"), "anthropic_workspace_member.test", tfjsonpath.New("workspace_id"), compare.ValuesSame()),
					statecheck.CompareValuePairs(rn, tfjsonpath.New("user_id"), "anthropic_workspace_member.test", tfjsonpath.New("user_id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("user_id"), knownvalue.StringExact(acctest.TestUserId)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("workspace_role"), knownvalue.StringExact("workspace_developer")),
				},
			},
		},
	})
}

func testAccWorkspaceMemberDataSourceConfig(workspaceName string) string {
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

data "anthropic_workspace_member" "test" {
	workspace_id = anthropic_workspace_member.test.workspace_id
	user_id      = anthropic_workspace_member.test.user_id
}
`, workspaceName, acctest.TestUserId)
}
