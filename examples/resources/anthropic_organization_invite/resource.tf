resource "anthropic_organization_invite" "developer" {
  email = "developer@example.com"
  role  = "developer"
}

resource "anthropic_organization_invite" "user" {
  email = "user@example.com"
  role  = "user"
}