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
	resource.AddTestSweepers("anthropic_environment", &resource.Sweeper{
		Name: "anthropic_environment",
		F: func(r string) error {
			ctx := context.Background()

			params := &apiclient.ListEnvironmentsParams{
				Limit: new(100),
			}

			for {
				httpResp, err := acctest.SharedClient.ListEnvironmentsWithResponse(ctx, params)
				if err != nil {
					return fmt.Errorf("Unable to list environments: %s", err)
				}
				if httpResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("Unable to list environments, got status %d: %s", httpResp.StatusCode(), string(httpResp.Body))
				}

				for _, env := range httpResp.JSON200.Data {
					if !strings.HasPrefix(env.Name, "tf-") {
						continue
					}

					log.Printf("[INFO] Destroying environment %s (%s)", env.Id, env.Name)

					// Reuse the production helper so the sweeper handles the
					// API's 409-conflict path the same way Delete does, and
					// surfaces non-2xx responses as errors instead of silently
					// continuing. The generated WithResponse helpers leave
					// HTTP 4xx/5xx in the response body with err == nil, so a
					// plain `if err != nil` check is not sufficient.
					if err := deleteEnvironmentWithFallback(ctx, acctest.SharedClient, env.Id); err != nil {
						log.Printf("[ERROR] Unable to clean up environment %s: %s", env.Id, err)
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

func TestAccEnvironmentResource(t *testing.T) {
	rn := "anthropic_environment.test"
	envName := acctest.RandomWithPrefix("tf-environment")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentResourceConfigUnrestricted(envName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(envName)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("config").AtMapKey("type"), knownvalue.StringExact("cloud")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("config").AtMapKey("networking").AtMapKey("type"), knownvalue.StringExact("unrestricted")),
				},
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEnvironmentResourceConfigLimited(envName + "-limited"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(envName+"-limited")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("config").AtMapKey("networking").AtMapKey("type"), knownvalue.StringExact("limited")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("config").AtMapKey("packages").AtMapKey("pip"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("pandas"),
					})),
				},
			},
		},
	})
}

func testAccEnvironmentResourceConfigUnrestricted(name string) string {
	return fmt.Sprintf(`
resource "anthropic_environment" "test" {
	name = %[1]q
	config = {
		networking = {
			type = "unrestricted"
		}
	}
}
`, name)
}

func testAccEnvironmentResourceConfigLimited(name string) string {
	return fmt.Sprintf(`
resource "anthropic_environment" "test" {
	name = %[1]q
	config = {
		networking = {
			type                   = "limited"
			allowed_hosts          = ["api.example.com"]
			allow_package_managers = true
		}
		packages = {
			pip = ["pandas"]
		}
	}
}
`, name)
}
