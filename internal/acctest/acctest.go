package acctest

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jianyuan/go-utils/must"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

var (
	TestApiKey = os.Getenv("ANTHROPIC_API_KEY")
	TestUserId = os.Getenv("ANTHROPIC_TEST_USER_ID")

	SharedClient *apiclient.ClientWithResponses
)

func init() {
	SharedClient = must.Get(apiclient.NewClientWithResponses(
		"https://api.anthropic.com",
		apiclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("anthropic-version", "2023-06-01")
			req.Header.Set("x-api-key", TestApiKey)
			return nil
		}),
		apiclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			for _, prefix := range []string{"/v1/agents", "/v1/environments", "/v1/vaults"} {
				if strings.HasPrefix(req.URL.Path, prefix) {
					req.Header.Set("anthropic-beta", "managed-agents-2026-04-01")
					return nil
				}
			}
			return nil
		}),
	))
}

func PreCheck(t *testing.T) {
	if TestApiKey == "" {
		t.Fatal("ANTHROPIC_API_KEY must be set for acceptance tests")
	}

	if TestUserId == "" {
		t.Fatal("ANTHROPIC_TEST_USER_ID must be set for acceptance tests")
	}
}
