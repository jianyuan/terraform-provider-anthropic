# Create a service account
resource "anthropic_service_account" "example" {
  name              = "ci-deploy-bot"
  organization_role = "developer"
}
