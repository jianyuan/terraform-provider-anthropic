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
	"github.com/jianyuan/terraform-provider-anthropic/internal/acctest"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

func init() {
	resource.AddTestSweepers("anthropic_vault", &resource.Sweeper{
		Name: "anthropic_vault",
		F: func(r string) error {
			ctx := context.Background()

			params := &apiclient.ListVaultsParams{
				Limit: new(100),
			}

			for {
				httpResp, err := acctest.SharedClient.ListVaultsWithResponse(ctx, params)
				if err != nil {
					return fmt.Errorf("Unable to list vaults: %s", err)
				}
				if httpResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("Unable to list vaults, got status %d: %s", httpResp.StatusCode(), string(httpResp.Body))
				}

				for _, vault := range httpResp.JSON200.Data {
					if !strings.HasPrefix(vault.DisplayName, "tf-") {
						continue
					}

					log.Printf("[INFO] Destroying vault %s (%s)", vault.Id, vault.DisplayName)

					// Use the production helper so non-2xx HTTP responses are
					// surfaced as errors instead of silently swallowed (the
					// generated WithResponse helper returns err == nil on
					// HTTP 4xx/5xx).
					if err := deleteVault(ctx, acctest.SharedClient, vault.Id); err != nil {
						log.Printf("[ERROR] Unable to clean up vault %s: %s", vault.Id, err)
					}
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

func TestAccVaultResource(t *testing.T) {
	rn := "anthropic_vault.test"
	displayName := acctest.RandomWithPrefix("tf-vault")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVaultResourceConfig(displayName, map[string]string{"env": "dev"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("display_name"), knownvalue.StringExact(displayName)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("metadata"), knownvalue.MapExact(map[string]knownvalue.Check{
						"env": knownvalue.StringExact("dev"),
					})),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("created_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("archived_at"), knownvalue.Null()),
				},
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVaultResourceConfig(displayName+"-updated", map[string]string{"env": "prod", "tier": "gold"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("display_name"), knownvalue.StringExact(displayName+"-updated")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("metadata"), knownvalue.MapExact(map[string]knownvalue.Check{
						"env":  knownvalue.StringExact("prod"),
						"tier": knownvalue.StringExact("gold"),
					})),
				},
			},
		},
	})
}

func testAccVaultResourceConfig(displayName string, metadata map[string]string) string {
	var meta strings.Builder
	if len(metadata) > 0 {
		meta.WriteString("metadata = {\n")
		for k, v := range metadata {
			fmt.Fprintf(&meta, "\t\t%q = %q\n", k, v)
		}
		meta.WriteString("\t}\n")
	}
	return fmt.Sprintf(`
resource "anthropic_vault" "test" {
	display_name = %[1]q
	%[2]s
}
`, displayName, meta.String())
}
