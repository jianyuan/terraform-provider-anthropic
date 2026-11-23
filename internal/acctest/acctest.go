package acctest

import (
	"os"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
)

// Environment variables read by acceptance tests. Pass the ones a test needs to PreCheck.
const (
	EnvApiKey    = "ANTHROPIC_API_KEY"
	EnvAuthToken = "ANTHROPIC_AUTH_TOKEN"
	EnvUserId    = "ANTHROPIC_TEST_USER_ID"
)

var (
	TestApiKey = os.Getenv(EnvApiKey)
	TestUserId = os.Getenv(EnvUserId)

	SharedApiKeyClient *anthropic.Client
)

func init() {
	SharedApiKeyClient = new(anthropic.NewClient(
		option.WithoutEnvironmentDefaults(),
		option.WithAPIKey(TestApiKey),
	))
}

// PreCheck fails the test unless every listed environment variable is set.
func PreCheck(t *testing.T, envs ...string) {
	for _, env := range envs {
		if os.Getenv(env) == "" {
			t.Fatalf("%s must be set for this acceptance test", env)
		}
	}
}
