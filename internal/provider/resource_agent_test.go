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
	resource.AddTestSweepers("anthropic_agent", &resource.Sweeper{
		Name: "anthropic_agent",
		F: func(r string) error {
			ctx := context.Background()

			params := &apiclient.ListAgentsParams{}

			for {
				httpResp, err := acctest.SharedClient.ListAgentsWithResponse(ctx, params)
				if err != nil {
					return fmt.Errorf("unable to list agents: %s", err)
				}

				if httpResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("unable to list agents, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body))
				}

				if httpResp.JSON200 == nil {
					break
				}

				for _, agent := range httpResp.JSON200.Data {
					if !strings.HasPrefix(agent.Name, "tf-") {
						continue
					}

					log.Printf("[INFO] Archiving agent %s", agent.Id)

					_, err := acctest.SharedClient.ArchiveAgentWithResponse(ctx, agent.Id, withManagedAgentsBeta)
					if err != nil {
						log.Printf("[ERROR] Unable to archive agent %s: %s", agent.Id, err)
						continue
					}

					log.Printf("[INFO] Archived agent %s", agent.Id)
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

func TestAccAgentResource_basic(t *testing.T) {
	rn := "anthropic_agent.test"
	agentName := acctest.RandomWithPrefix("tf-agent")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheckManagedAgents(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfig_basic(agentName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(agentName)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("model"), knownvalue.StringExact("claude-sonnet-4-5")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("version"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAgentResourceConfig_basic(agentName + "-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(agentName+"-updated")),
				},
			},
		},
	})
}

func TestAccAgentResource_withTools(t *testing.T) {
	rn := "anthropic_agent.test"
	agentName := acctest.RandomWithPrefix("tf-agent")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheckManagedAgents(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfig_withTools(agentName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(agentName)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("system"), knownvalue.StringExact("You are a helpful assistant.")),
				},
			},
		},
	})
}

func TestAccAgentResource_withSystem(t *testing.T) {
	rn := "anthropic_agent.test"
	agentName := acctest.RandomWithPrefix("tf-agent")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheckManagedAgents(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfig_withSystem(agentName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("system"), knownvalue.StringExact("You are a DevOps assistant.")),
				},
			},
		},
	})
}

// Regression test: skills must round-trip the JSON field name `skill_id`
// (not `id`) when sent to the API. Previously the OpenAPI schema for
// AgentSkillRequest mapped it to `id`, which the API rejected with
// "skills.0.id: Extra inputs are not permitted".
func TestAccAgentResource_withSkill(t *testing.T) {
	rn := "anthropic_agent.test"
	agentName := acctest.RandomWithPrefix("tf-agent")
	skillName := strings.ToLower(acctest.RandomWithPrefix("tf-skill"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheckManagedAgents(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfig_withSkill(agentName, skillName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("skills"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("skills").AtSliceIndex(0).AtMapKey("type"), knownvalue.StringExact("custom")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("skills").AtSliceIndex(0).AtMapKey("version"), knownvalue.StringExact("latest")),
				},
			},
		},
	})
}

func testAccAgentResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "anthropic_agent" "test" {
	name  = %[1]q
	model = "claude-sonnet-4-5"
}
`, name)
}

func testAccAgentResourceConfig_withTools(name string) string {
	return fmt.Sprintf(`
resource "anthropic_agent" "test" {
	name   = %[1]q
	model  = "claude-sonnet-4-5"
	system = "You are a helpful assistant."

	tools {
		type = "agent_toolset_20251212"
	}
}
`, name)
}

func testAccAgentResourceConfig_withSystem(name string) string {
	return fmt.Sprintf(`
resource "anthropic_agent" "test" {
	name   = %[1]q
	model  = "claude-sonnet-4-5"
	system = "You are a DevOps assistant."
}
`, name)
}

// Regression test: tools[].default_config.permission_policy must round-trip
// through the API. Without this, the API silently re-applies the default
// `always_ask` policy on every TF apply, breaking unattended sessions.
func TestAccAgentResource_withToolDefaultConfig(t *testing.T) {
	rn := "anthropic_agent.test"
	agentName := acctest.RandomWithPrefix("tf-agent")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheckManagedAgents(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfig_withToolDefaultConfig(agentName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tools").AtSliceIndex(0).AtMapKey("default_config").AtMapKey("enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tools").AtSliceIndex(0).AtMapKey("default_config").AtMapKey("permission_policy").AtMapKey("type"), knownvalue.StringExact("always_allow")),
				},
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAgentResourceConfig_withToolDefaultConfig(name string) string {
	return fmt.Sprintf(`
resource "anthropic_agent" "test" {
	name  = %[1]q
	model = "claude-sonnet-4-5"

	tools {
		type = "agent_toolset_20251212"

		default_config {
			enabled = true
			permission_policy {
				type = "always_allow"
			}
		}
	}
}
`, name)
}

func testAccAgentResourceConfig_withSkill(agentName, skillName string) string {
	return fmt.Sprintf(`
resource "anthropic_skill" "test" {
	display_title = %[2]q
	skill_name    = %[2]q
	content       = "---\nname: %[2]s\ndescription: Test skill for agent attachment\n---\n\nTest content."
}

resource "anthropic_agent" "test" {
	name  = %[1]q
	model = "claude-sonnet-4-5"

	skills {
		skill_id = anthropic_skill.test.id
		type     = "custom"
		version  = "latest"
	}
}
`, agentName, skillName)
}
