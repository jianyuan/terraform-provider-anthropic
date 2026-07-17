# Manual-only deployment: no schedule block, triggered via `ant beta:deployments run`
# or the API. Floats to the agent's latest version (agent_version left unset).
resource "anthropic_deployment" "manual" {
  name           = "readme-updater manual trigger"
  description    = "Runs the readme-updater agent on demand; no automatic schedule."
  agent_id       = anthropic_agent.readme_updater.id
  environment_id = anthropic_environment.readme_updater.id
  vault_ids      = [var.github_vault_id]

  initial_events = [jsonencode({
    type    = "user.message"
    content = [{ type = "text", text = "frank-bee/readme-agent-demo" }]
  })]

  resources = [jsonencode({
    type            = "memory_store"
    memory_store_id = var.memory_store_id
    access          = "read_write"
    instructions    = "Read/update state.md to track the last commit SHA you have documented."
  })]
}

# Scheduled deployment: runs automatically on a weekly cron, pinned to a specific
# agent version.
resource "anthropic_deployment" "scheduled" {
  name           = "ci-triage weekly report"
  agent_id       = anthropic_agent.ci_triage.id
  agent_version  = "4"
  environment_id = anthropic_environment.ci_triage.id

  schedule {
    expression = "0 3 * * 0"
    timezone   = "UTC"
  }
}
