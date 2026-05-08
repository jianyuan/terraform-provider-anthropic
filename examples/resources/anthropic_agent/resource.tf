# Minimal agent: name + model only.
resource "anthropic_agent" "minimal" {
  name = "Minimal Assistant"
  model = {
    id = "claude-opus-4-7"
  }
  system = "You are a helpful assistant."
}

# Coding agent with the pre-built agent toolset, a custom tool, and metadata.
resource "anthropic_agent" "coder" {
  name        = "Coding Assistant"
  description = "Pair programmer with web search disabled."
  model = {
    id = "claude-opus-4-7"
  }
  system = "You are an expert software engineer. Always write tests."

  tools = [
    {
      type = "agent_toolset_20260401"
      configs = [
        { name = "web_fetch", enabled = false },
        { name = "web_search", enabled = false },
      ]
    },
    {
      type        = "custom"
      name        = "lookup_ticket"
      description = "Fetch the title and description of a Linear ticket by identifier."
      # input_schema is a JSON-encoded JSON Schema. Use jsonencode(...) so the
      # provider can semantically compare the value across refreshes; nesting
      # this attribute inside a list is a Terraform Plugin Framework
      # constraint that prevents using a native-HCL DynamicAttribute here.
      input_schema = jsonencode({
        type = "object"
        properties = {
          identifier = {
            type        = "string"
            description = "Linear identifier such as ENG-123."
          }
        }
        required = ["identifier"]
      })
    },
  ]

  metadata = {
    team = "platform"
    env  = "prod"
  }
}

# Agent with an MCP server and skills.
resource "anthropic_agent" "github_assistant" {
  name = "GitHub Assistant"
  model = {
    id = "claude-opus-4-7"
  }

  mcp_servers = [
    {
      type = "url"
      name = "github"
      url  = "https://api.githubcopilot.com/mcp/"
    },
  ]

  tools = [
    { type = "agent_toolset_20260401" },
    { type = "mcp_toolset", mcp_server_name = "github" },
  ]

  skills = [
    { type = "anthropic", skill_id = "xlsx" },
  ]
}
