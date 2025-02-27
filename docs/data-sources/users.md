---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "anthropic_users Data Source - terraform-provider-anthropic"
subcategory: ""
description: |-
  List all users in the Organization.
---

# anthropic_users (Data Source)

List all users in the Organization.

## Example Usage

```terraform
data "anthropic_users" "example" {
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `users` (Attributes Set) List of users. (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `added_at` (String) RFC 3339 datetime string indicating when the User joined the Organization.
- `email` (String) Email of the User.
- `id` (String) ID of the User.
- `name` (String) Name of the User.
- `role` (String) Organization role of the User.
