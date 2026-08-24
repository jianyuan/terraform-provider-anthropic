package acctest

import (
	"context"
	"net/http"
	"os"
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
