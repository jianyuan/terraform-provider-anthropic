data "anthropic_organization_invites" "all" {}

output "pending_invites" {
  description = "All pending organization invites"
  value       = data.anthropic_organization_invites.all.invites
}

output "invite_count" {
  description = "Number of pending invites"
  value       = length(data.anthropic_organization_invites.all.invites)
}