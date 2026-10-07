package provider

import (
	"context"
	"errors"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
)

type base struct {
	apiKey          string
	authToken       string
	apiKeyClient    *anthropic.Client
	authTokenClient *anthropic.Client
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
