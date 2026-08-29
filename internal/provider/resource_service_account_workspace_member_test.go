package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
)

func TestAccServiceAccountWorkspaceMemberResource(t *testing.T) {
	rn := "anthropic_service_account_workspace_member.test"
	name := sdkacctest.RandomWithPrefix("tf-sawm")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountWorkspaceMemberResourceConfig(name, "workspace_admin"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("workspace_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("service_account_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("workspace_role"), knownvalue.StringExact("workspace_admin")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("created_by_actor_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("implicit"), knownvalue.Null()),
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
					serviceAccountId := rs.Primary.Attributes["service_account_id"]
					return BuildTwoPartId(workspaceId, serviceAccountId), nil
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "service_account_id",
			},
			{
				Config: testAccServiceAccountWorkspaceMemberResourceConfig(name, "workspace_restricted_developer"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("workspace_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("service_account_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("workspace_role"), knownvalue.StringExact("workspace_restricted_developer")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("created_by_actor_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("implicit"), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccServiceAccountWorkspaceMemberResourceConfig(name, workspaceRole string) string {
	return testAccWorkspaceResourceConfig(name, "") + testAccServiceAccountResourceConfig(name, "developer", "") + fmt.Sprintf(`
resource "anthropic_service_account_workspace_member" "test" {
	workspace_id = anthropic_workspace.test.id
	service_account_id = anthropic_service_account.test.id
	workspace_role = "%s"
}
`, workspaceRole)
}
