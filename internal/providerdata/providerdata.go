package providerdata

import "github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"

type ProviderData struct {
	ApiKey    string
	AuthToken string
	Client    *apiclient.ClientWithResponses
}
