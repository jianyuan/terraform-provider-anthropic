package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
)

func TestAccWorkspaceMemberResource(t *testing.T) {
	rn := "anthropic_workspace_member.test"
	workspaceName := acctest.RandomWithPrefix("tf-workspace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceMemberResourceConfig(workspaceName, "workspace_user"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(rn, tfjsonpath.New("workspace_id"), "anthropic_workspace.test", tfjsonpath.New("id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("user_id"), knownvalue.StringExact(acctest.TestUserId)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("workspace_role"), knownvalue.StringExact("workspace_user")),
				},
			},
			{
				ResourceName: rn,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[rn]
					if !ok {
						return "", fmt.Errorf("not found: %s", rn)
					}
					workspaceId := rs.Primary.Attributes["workspace_id"]
					userId := rs.Primary.Attributes["user_id"]
					return BuildTwoPartId(workspaceId, userId), nil
				},
			},
			{
				Config: testAccWorkspaceMemberResourceConfig(workspaceName, "workspace_developer"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(rn, tfjsonpath.New("workspace_id"), "anthropic_workspace.test", tfjsonpath.New("id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("user_id"), knownvalue.StringExact(acctest.TestUserId)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("workspace_role"), knownvalue.StringExact("workspace_developer")),
				},
			},
		},
	})
}

func testAccWorkspaceMemberResourceConfig(workspaceName string, workspaceRole string) string {
	return fmt.Sprintf(`
resource "anthropic_workspace" "test" {
	name = %[1]q
}

resource "anthropic_workspace_member" "test" {
	workspace_id   = anthropic_workspace.test.id
	user_id        = %[2]q
	workspace_role = %[3]q
}
`, workspaceName, acctest.TestUserId, workspaceRole)
}
