package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
)

func init() {
	resource.AddTestSweepers("anthropic_workspace", &resource.Sweeper{
		Name: "anthropic_workspace",
		F: func(r string) error {
			ctx := context.Background()

			iter := acctest.SharedApiKeyClient.Organization.Workspaces.ListAutoPaging(ctx, anthropic.OrganizationWorkspaceListParams{})

			for iter.Next() {
				workspace := iter.Current()

				if !strings.HasPrefix(workspace.Name, "tf-") {
					continue
				}

				log.Printf("[INFO] Destroying workspace %s", workspace.ID)

				_, err := acctest.SharedApiKeyClient.Organization.Workspaces.Archive(
					ctx,
					workspace.ID,
				)

				if err != nil {
					log.Printf("[ERROR] Unable to archive workspace %s: %s", workspace.ID, err)
					continue
				}

				log.Printf("[INFO] Archived workspace %s", workspace.ID)
			}

			if err := iter.Err(); err != nil {
				log.Printf("[ERROR] Unable to list workspaces: %s", err)
			}

			return nil
		},
	})
}

func TestAccWorkspaceResource(t *testing.T) {
	rn := "anthropic_workspace.test"
	workspaceName := sdkacctest.RandomWithPrefix("tf-workspace")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceResourceConfig(workspaceName, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(workspaceName)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("created_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("archived_at"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("display_color"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("compartment_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("data_residency"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"allowed_inference_geos": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"values":       knownvalue.Null(),
							"unrestricted": knownvalue.Bool(true),
						}),
						"default_inference_geo": knownvalue.StringExact("global"),
						"workspace_geo":         knownvalue.StringExact("us"),
					})),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("external_key_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tags"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceResourceConfig(workspaceName+"-updated", `
					tags = {
						foo = "bar"
					}
				`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(workspaceName+"-updated")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("created_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("archived_at"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("display_color"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("compartment_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("data_residency"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"allowed_inference_geos": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"values":       knownvalue.Null(),
							"unrestricted": knownvalue.Bool(true),
						}),
						"default_inference_geo": knownvalue.StringExact("global"),
						"workspace_geo":         knownvalue.StringExact("us"),
					})),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("external_key_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tags"), knownvalue.MapExact(map[string]knownvalue.Check{
						"foo": knownvalue.StringExact("bar"),
					})),
				},
			},
		},
	})
}

func testAccWorkspaceResourceConfig(name, extra string) string {
	return fmt.Sprintf(`
resource "anthropic_workspace" "test" {
	name = %[1]q
	%[2]s
}
`, name, extra)
}
