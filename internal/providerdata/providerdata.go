package providerdata

import (
	"github.com/anthropics/anthropic-sdk-go/v2"
)

type ProviderData struct {
	ApiKey          string
	AuthToken       string
	ApiKeyClient    *anthropic.Client
	AuthTokenClient *anthropic.Client
}
