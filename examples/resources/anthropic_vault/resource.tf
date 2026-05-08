# A vault is a per-end-user collection of MCP credentials. The credentials
# themselves are not managed by this resource and must be created via the
# API or SDK out-of-band.
resource "anthropic_vault" "alice" {
  display_name = "Alice"

  metadata = {
    external_user_id = "usr_abc123"
  }
}
