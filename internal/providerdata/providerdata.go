package providerdata

import (
	"github.com/anthropics/anthropic-sdk-go"
)

type ProviderData struct {
	ApiKey          string
	AuthToken       string
	ApiKeyClient    *anthropic.Client
	AuthTokenClient *anthropic.Client
}
