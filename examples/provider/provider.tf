# Configure the Anthropic provider
provider "anthropic" {
  api_key = "sk-ant-adminxx-xxxxx-xxxxx-xxxxx-xxxxx"
}

# Create a new workspace
resource "anthropic_workspace" "example" {
  name = "Workspace Name"
}

# Retrieve a user by ID
data "anthropic_user" "example" {
  id = "user_xxxxx"
}

# Add a user to the workspace
resource "anthropic_workspace_member" "example" {
  workspace_id   = anthropic_workspace.example.id
  user_id        = data.anthropic_user.example.id
  workspace_role = "workspace_developer"
}
