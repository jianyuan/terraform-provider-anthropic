package acctest

import (
	"os"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

var (
	TestApiKey = os.Getenv("ANTHROPIC_API_KEY")
	TestUserId = os.Getenv("ANTHROPIC_TEST_USER_ID")

	SharedApiKeyClient *anthropic.Client
)

func init() {
	SharedApiKeyClient = new(anthropic.NewClient(
		option.WithoutEnvironmentDefaults(),
		option.WithAPIKey(TestApiKey),
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
