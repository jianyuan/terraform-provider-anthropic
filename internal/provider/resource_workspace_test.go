package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

func init() {
	resource.AddTestSweepers("anthropic_workspace", &resource.Sweeper{
		Name: "anthropic_workspace",
		F: func(r string) error {
			ctx := context.Background()

			params := &apiclient.ListWorkspacesParams{
				Limit: new(int64(100)),
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

				if !httpResp.JSON200.HasMore || httpResp.JSON200.LastId.IsNull() || !httpResp.JSON200.LastId.IsSpecified() {
					break
				}

				params.AfterId = new(httpResp.JSON200.LastId.MustGet())
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
						"allowed_inference_geos":              knownvalue.Null(),
						"allowed_inference_geos_unrestricted": knownvalue.Bool(true),
						"default_inference_geo":               knownvalue.StringExact("global"),
						"workspace_geo":                       knownvalue.StringExact("us"),
					})),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("external_key_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tags"), knownvalue.MapExact(map[string]knownvalue.Check{
						"foo": knownvalue.StringExact("bar"),
					})),
				},
			},
			{
				Config: testAccWorkspaceResourceConfig(workspaceName+"-updated", `
					data_residency = {
						allowed_inference_geos = ["us"]
						default_inference_geo = "us"
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
						"allowed_inference_geos": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("us"),
						}),
						"allowed_inference_geos_unrestricted": knownvalue.Null(),
						"default_inference_geo":               knownvalue.StringExact("us"),
						"workspace_geo":                       knownvalue.StringExact("us"),
					})),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("external_key_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tags"), knownvalue.MapSizeExact(0)),
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
