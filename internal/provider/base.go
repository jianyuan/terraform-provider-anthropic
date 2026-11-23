package provider

import (
	"context"
	"errors"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	requiresApiKeyNote    = "\n\nRequires an Admin API key: set `api_key` in the provider block or the `ANTHROPIC_API_KEY` environment variable. OAuth bearer tokens (`auth_token`) are not used."
	requiresAuthTokenNote = "\n\nRequires an OAuth bearer token: set `auth_token` in the provider block or the `ANTHROPIC_AUTH_TOKEN` environment variable. Admin API keys (`api_key`) are not accepted."
)

type base struct {
	// Credentials checked at Configure time; each resource and data source sets the one its endpoints accept.
	requiresApiKey    bool
	requiresAuthToken bool

	apiKey          string
	authToken       string
	apiKeyClient    *anthropic.Client
	authTokenClient *anthropic.Client
}

// checkCredentials reports a missing credential when the resource or data source is configured.
func (b *base) checkCredentials(diags *diag.Diagnostics) {
	if b.requiresAuthToken && b.authToken == "" {
		diags.AddError(
			"Missing auth_token",
			"This resource or data source requires an OAuth bearer token. Set `auth_token` in the provider block or the ANTHROPIC_AUTH_TOKEN environment variable. Admin API keys (api_key / ANTHROPIC_API_KEY) are not accepted here.",
		)
	}
	if b.requiresApiKey && b.apiKey == "" {
		diags.AddError(
			"Missing api_key",
			"This resource or data source requires an Admin API key. Set `api_key` in the provider block or the ANTHROPIC_API_KEY environment variable. OAuth bearer tokens (auth_token / ANTHROPIC_AUTH_TOKEN) are not used here.",
		)
	}
}

func (b *base) WithApiKeyRequestEditorFn() func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		if b.apiKey == "" {
			return errors.New("api_key is required")
		}
		req.Header.Set("x-api-key", b.apiKey)
		return nil
	}
}

func (b *base) WithAuthTokenRequestEditorFn() func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		if b.authToken == "" {
			return errors.New("auth_token is required")
		}
		req.Header.Set("authorization", "Bearer "+b.authToken)
		return nil
	}
}
