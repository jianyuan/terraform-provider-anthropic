---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "anthropic_workspace_member Data Source - terraform-provider-anthropic"
subcategory: ""
description: |-
  Get a member of a Workspace.
---

# anthropic_workspace_member (Data Source)

Get a member of a Workspace.

## Example Usage

```terraform
data "anthropic_workspace_member" "example" {
  workspace_id = "wrkspc_xxxxx"
  user_id      = "user_xxxxx"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `user_id` (String) ID of the user who is a member of the Workspace.
- `workspace_id` (String) ID of the Workspace to which the member belongs.

### Read-Only

- `workspace_role` (String) Role of the new Workspace Member. Must be one of `workspace_user`, `workspace_developer`, or `workspace_admin`.