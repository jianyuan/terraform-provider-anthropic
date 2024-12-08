resource "anthropic_workspace" "example" {
  name = "Workspace Name"
}

data "anthropic_user" "example" {
  id = "user_xxxxx"
}

# Create a workspace member
resource "anthropic_workspace_member" "example" {
  workspace_id   = anthropic_workspace.example.id
  user_id        = data.anthropic_user.example.id
  workspace_role = "workspace_developer"
}
