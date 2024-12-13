---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "anthropic Provider"
subcategory: ""
description: |-
  
---

# anthropic Provider



## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive) The Admin API key for authentication. Get this from the [Anthropic console](https://console.anthropic.com/settings/admin-keys). It can be sourced from the `ANTHROPIC_API_KEY` environment variable.
- `base_url` (String) API endpoint for the Anthropic service. Defaults to `https://api.anthropic.com`. It can be sourced from the `ANTHROPIC_BASE_URL` environment variable.
