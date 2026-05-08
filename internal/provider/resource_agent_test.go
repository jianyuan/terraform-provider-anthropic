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
	resource.AddTestSweepers("anthropic_agent", &resource.Sweeper{
		Name: "anthropic_agent",
		F: func(r string) error {
			ctx := context.Background()

			params := &apiclient.ListAgentsParams{
				Limit: new(100),
			}

			for {
				httpResp, err := acctest.SharedClient.ListAgentsWithResponse(ctx, params)
				if err != nil {
					return fmt.Errorf("Unable to list agents: %s", err)
				}
				if httpResp.StatusCode() != http.StatusOK {
					return fmt.Errorf("Unable to list agents, got status %d: %s", httpResp.StatusCode(), string(httpResp.Body))
				}

				for _, agent := range httpResp.JSON200.Data {
					if !strings.HasPrefix(agent.Name, "tf-") {
						continue
					}

					log.Printf("[INFO] Archiving agent %s (%s)", agent.Id, agent.Name)

					archiveResp, err := acctest.SharedClient.ArchiveAgentWithResponse(ctx, agent.Id)
					if err != nil {
						log.Printf("[ERROR] Unable to archive agent %s: %s", agent.Id, err)
						continue
					}
					// The generated WithResponse helper returns err == nil on
					// HTTP 4xx/5xx; check the status code so a failed archive
					// doesn't silently leak resources.
					if code := archiveResp.StatusCode(); code != http.StatusOK && code != http.StatusNotFound {
						log.Printf("[ERROR] Archive of agent %s returned status %d: %s", agent.Id, code, string(archiveResp.Body))
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

func TestAccAgentResource_basic(t *testing.T) {
	rn := "anthropic_agent.test"
	name := acctest.RandomWithPrefix("tf-agent")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfigBasic(name, "You are a helpful assistant.", map[string]string{"env": "dev"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("model").AtMapKey("id"), knownvalue.StringExact("claude-opus-4-7")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("system"), knownvalue.StringExact("You are a helpful assistant.")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("version"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("metadata"), knownvalue.MapExact(map[string]knownvalue.Check{
						"env": knownvalue.StringExact("dev"),
					})),
				},
			},
			{
				ResourceName:                         rn,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
			{
				Config: testAccAgentResourceConfigBasic(name, "You are a helpful assistant. Always write tests.", map[string]string{"env": "prod", "team": "platform"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("system"), knownvalue.StringExact("You are a helpful assistant. Always write tests.")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("version"), knownvalue.Int64Func(func(v int64) error {
						if v <= 1 {
							return fmt.Errorf("expected version > 1, got %d", v)
						}
						return nil
					})),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("metadata"), knownvalue.MapExact(map[string]knownvalue.Check{
						"env":  knownvalue.StringExact("prod"),
						"team": knownvalue.StringExact("platform"),
					})),
				},
			},
		},
	})
}

func TestAccAgentResource_tools(t *testing.T) {
	rn := "anthropic_agent.test"
	name := acctest.RandomWithPrefix("tf-agent")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfigTools(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tools").AtSliceIndex(0).AtMapKey("type"), knownvalue.StringExact("agent_toolset_20260401")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tools").AtSliceIndex(1).AtMapKey("type"), knownvalue.StringExact("custom")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("tools").AtSliceIndex(1).AtMapKey("name"), knownvalue.StringExact("get_weather")),
				},
			},
		},
	})
}

func testAccAgentResourceConfigBasic(name, system string, metadata map[string]string) string {
	var meta strings.Builder
	if len(metadata) > 0 {
		meta.WriteString("metadata = {\n")
		for k, v := range metadata {
			fmt.Fprintf(&meta, "\t\t%q = %q\n", k, v)
		}
		meta.WriteString("\t}\n")
	}
	return fmt.Sprintf(`
resource "anthropic_agent" "test" {
	name   = %[1]q
	model  = { id = "claude-opus-4-7" }
	system = %[2]q
	%[3]s
}
`, name, system, meta.String())
}

func testAccAgentResourceConfigTools(name string) string {
	return fmt.Sprintf(`
resource "anthropic_agent" "test" {
	name  = %[1]q
	model = { id = "claude-opus-4-7" }
	tools = [
		{
			type = "agent_toolset_20260401"
			configs = [
				{ name = "web_fetch", enabled = false },
			]
		},
		{
			type        = "custom"
			name        = "get_weather"
			description = "Get current weather for a location."
			input_schema = jsonencode({
				type = "object"
				properties = {
					location = { type = "string", description = "City name" }
				}
				required = ["location"]
			})
		},
	]
}
`, name)
}
