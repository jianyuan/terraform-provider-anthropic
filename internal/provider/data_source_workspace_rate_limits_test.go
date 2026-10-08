package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
)

func TestAccWorkspaceRateLimitsDataSource(t *testing.T) {
	rn := "data.anthropic_workspace_rate_limits.test"
	workspaceName := sdkacctest.RandomWithPrefix("tf-workspace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceRateLimitsDataSourceConfig(workspaceName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(rn, tfjsonpath.New("workspace_id"), "anthropic_workspace.test", tfjsonpath.New("id"), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccWorkspaceRateLimitsDataSourceConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "anthropic_workspace" "test" {
	name = %[1]q
}

data "anthropic_workspace_rate_limits" "test" {
	workspace_id = anthropic_workspace.test.id
	query = {
		include_inherited = true
	}
}

check "workspace_ids" {
	assert {
		condition = alltrue([
			for limit in data.anthropic_workspace_rate_limits.test.workspace_rate_limits :
			limit.workspace_id == anthropic_workspace.test.id
		])
		error_message = "One or more rate limit entries contain an unexpected workspace_id."
	}
}
`, workspaceName)
}
