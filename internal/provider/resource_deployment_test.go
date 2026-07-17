package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/frank-bee/terraform-provider-anthropic/internal/acctest"
	"github.com/frank-bee/terraform-provider-anthropic/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func init() {
	// Deployments have no hard delete — the sweeper archives leftovers instead.
	resource.AddTestSweepers("anthropic_deployment", &resource.Sweeper{
		Name: "anthropic_deployment",
		F: func(r string) error {
			ctx := context.Background()

			params := &apiclient.ListDeploymentsParams{}

			for {
				httpResp, err := acctest.SharedClient.ListDeploymentsWithResponse(ctx, params, withDeploymentsBeta)
				if err != nil {
					return fmt.Errorf("unable to list deployments: %s", err)
				}

				if httpResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("unable to list deployments, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body))
				}

				if httpResp.JSON200 == nil {
					break
				}

				for _, dep := range httpResp.JSON200.Data {
					if !strings.HasPrefix(dep.Name, "tf-") {
						continue
					}

					log.Printf("[INFO] Archiving deployment %s", dep.Id)

					_, err := acctest.SharedClient.ArchiveDeploymentWithResponse(ctx, dep.Id, withDeploymentsBeta)
					if err != nil {
						log.Printf("[ERROR] Unable to archive deployment %s: %s", dep.Id, err)
						continue
					}

					log.Printf("[INFO] Archived deployment %s", dep.Id)
				}

				if httpResp.JSON200.NextPage == nil || *httpResp.JSON200.NextPage == "" {
					break
				}
				params.Page = httpResp.JSON200.NextPage
			}

			return nil
		},
	})
}

// TestAccDeploymentResource_basic covers the happy path: create (manual, no
// schedule) against a fixture agent+environment, import, update name.
func TestAccDeploymentResource_basic(t *testing.T) {
	rn := "anthropic_deployment.test"
	depName := acctest.RandomWithPrefix("tf-deployment")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheckManagedAgents(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentResourceConfig_basic(depName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(depName)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("agent_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("agent_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("status"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("created_at"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeploymentResourceConfig_basic(depName + "-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(depName+"-updated")),
				},
			},
		},
	})
}

// TestAccDeploymentResource_schedule covers adding then removing a `schedule`
// block, confirming the resource clears it server-side rather than leaving a
// stale schedule behind (see deploymentUpdateBody in resource_deployment.go).
func TestAccDeploymentResource_schedule(t *testing.T) {
	rn := "anthropic_deployment.test"
	depName := acctest.RandomWithPrefix("tf-deployment")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheckManagedAgents(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentResourceConfig_schedule(depName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("schedule").AtMapKey("expression"), knownvalue.StringExact("0 3 * * 0")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("schedule").AtMapKey("timezone"), knownvalue.StringExact("UTC")),
				},
			},
			{
				Config: testAccDeploymentResourceConfig_basic(depName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("schedule"), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccDeploymentResourceFixtures(name string) string {
	return fmt.Sprintf(`
resource "anthropic_agent" "test" {
	name  = %[1]q
	model = "claude-sonnet-4-5"
}

resource "anthropic_environment" "test" {
	name            = %[1]q
	networking_type = "unrestricted"
}
`, name)
}

func testAccDeploymentResourceConfig_basic(name string) string {
	return testAccDeploymentResourceFixtures(name) + fmt.Sprintf(`
resource "anthropic_deployment" "test" {
	name           = %[1]q
	agent_id       = anthropic_agent.test.id
	environment_id = anthropic_environment.test.id

	initial_events = [jsonencode({
		type    = "user.message"
		content = [{ type = "text", text = "hello" }]
	})]
}
`, name)
}

func testAccDeploymentResourceConfig_schedule(name string) string {
	return testAccDeploymentResourceFixtures(name) + fmt.Sprintf(`
resource "anthropic_deployment" "test" {
	name           = %[1]q
	agent_id       = anthropic_agent.test.id
	environment_id = anthropic_environment.test.id

	initial_events = [jsonencode({
		type    = "user.message"
		content = [{ type = "text", text = "hello" }]
	})]

	schedule {
		expression = "0 3 * * 0"
		timezone   = "UTC"
	}
}
`, name)
}
