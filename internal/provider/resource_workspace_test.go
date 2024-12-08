package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/go-utils/ptr"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

func init() {
	resource.AddTestSweepers("anthropic_workspace", &resource.Sweeper{
		Name: "anthropic_workspace",
		F: func(r string) error {
			ctx := context.Background()

			params := &apiclient.ListWorkspacesParams{
				Limit: ptr.Ptr(100),
			}

			for {
				httpResp, err := acctest.SharedClient.ListWorkspacesWithResponse(
					ctx,
					params,
				)
				if err != nil {
					return fmt.Errorf("Unable to read, got error: %s", err)
				}

				if httpResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("Unable to read, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body))
				}

				for _, workspace := range httpResp.JSON200.Data {
					if !strings.HasPrefix(workspace.Name, "tf-") {
						continue
					}

					log.Printf("[INFO] Destroying workspace %s", workspace.Id)

					_, err := acctest.SharedClient.ArchiveWorkspaceWithResponse(
						ctx,
						workspace.Id,
					)

					if err != nil {
						log.Printf("[ERROR] Unable to archive workspace %s: %s", workspace.Id, err)
						continue
					}

					log.Printf("[INFO] Archived workspace %s", workspace.Id)
				}

				if !httpResp.JSON200.HasMore || httpResp.JSON200.LastId == nil {
					break
				}

				params.AfterId = httpResp.JSON200.LastId
			}

			return nil
		},
	})
}

func TestAccWorkspaceResource(t *testing.T) {
	rn := "anthropic_workspace.test"
	workspaceName := acctest.RandomWithPrefix("tf-workspace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceResourceConfig(workspaceName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(workspaceName)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("created_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("archived_at"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("display_color"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceResourceConfig(workspaceName + "-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(workspaceName+"-updated")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("created_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("archived_at"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("display_color"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccWorkspaceResourceConfig(workspaceName string) string {
	return fmt.Sprintf(`
resource "anthropic_workspace" "test" {
	name = %[1]q
}
`, workspaceName)
}
