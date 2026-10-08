data "anthropic_workspace_rate_limits" "example" {
  workspace_id = "wrkspc_xxxxx"

  # Optional query parameters to filter the results.
  query = {
    include_inherited = true
  }
}
