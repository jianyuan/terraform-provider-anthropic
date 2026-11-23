package provider

import (
	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/providerdata"
)

const (
	requiresApiKeyNote    = "\n\nRequires an Admin API key: set `api_key` in the provider block or the `ANTHROPIC_API_KEY` environment variable. OAuth bearer tokens (`auth_token`) are not used."
	requiresAuthTokenNote = "\n\nRequires an OAuth bearer token: set `auth_token` in the provider block or the `ANTHROPIC_AUTH_TOKEN` environment variable. Admin API keys (`api_key`) are not accepted."
)

type base struct {
	// Exactly one must be set; it picks the credential, and so the client, the endpoints accept.
	requiresApiKey    bool
	requiresAuthToken bool

	client *anthropic.Client
}

// configure checks the required credential is present and selects the matching client.
func (b *base) configure(providerData *providerdata.ProviderData, diags *diag.Diagnostics) {
	switch {
	case b.requiresApiKey == b.requiresAuthToken:
		diags.AddError(
			"Invalid credential requirement",
			"Exactly one of requiresApiKey or requiresAuthToken must be set. Please report this issue to the provider developers.",
		)
	case b.requiresApiKey && providerData.ApiKey == "":
		diags.AddError(
			"Missing api_key",
			"This resource or data source requires an Admin API key. Set `api_key` in the provider block or the ANTHROPIC_API_KEY environment variable. OAuth bearer tokens (auth_token / ANTHROPIC_AUTH_TOKEN) are not used here.",
		)
	case b.requiresAuthToken && providerData.AuthToken == "":
		diags.AddError(
			"Missing auth_token",
			"This resource or data source requires an OAuth bearer token. Set `auth_token` in the provider block or the ANTHROPIC_AUTH_TOKEN environment variable. Admin API keys (api_key / ANTHROPIC_API_KEY) are not accepted here.",
		)
	case b.requiresApiKey:
		b.client = providerData.ApiKeyClient
	default:
		b.client = providerData.AuthTokenClient
	}
}
