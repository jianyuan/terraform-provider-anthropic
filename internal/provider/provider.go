package provider

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/providerdata"
)

// Ensure AnthropicProvider satisfies various provider interfaces.
var _ provider.Provider = &AnthropicProvider{}
var _ provider.ProviderWithFunctions = &AnthropicProvider{}

// AnthropicProvider defines the provider implementation.
type AnthropicProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AnthropicProviderModel describes the provider data model.
type AnthropicProviderModel struct {
	BaseUrl   types.String `tfsdk:"base_url"`
	ApiKey    types.String `tfsdk:"api_key"`
	AuthToken types.String `tfsdk:"auth_token"`
}

func (p *AnthropicProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "anthropic"
	resp.Version = p.version
}

func (p *AnthropicProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Anthropic provider is used to interact with the Anthropic service.\n\nIf you find this provider useful, please consider supporting me through GitHub Sponsorship or Ko-Fi to help with its development.\n\n[![Github-sponsors](https://img.shields.io/badge/sponsor-30363D?style=for-the-badge&logo=GitHub-Sponsors&logoColor=#EA4AAA)](https://github.com/sponsors/jianyuan)\n[![Ko-Fi](https://img.shields.io/badge/Ko--fi-F16061?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ko-fi.com/L3L71DQEL)",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "API endpoint for the Anthropic service. Defaults to `https://api.anthropic.com`. It can be sourced from the `ANTHROPIC_BASE_URL` environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The Admin API key for authentication. Get this from the [Anthropic console](https://console.anthropic.com/settings/admin-keys). It can be sourced from the `ANTHROPIC_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "The OAuth bearer token for authentication. Log in with the `ant` CLI under a dedicated profile with the `org:admin` scope (see [Admin access](https://platform.claude.com/docs/en/cli-sdks-libraries/cli/authentication#admin-access)), then export the bearer token It can be sourced from the `ANTHROPIC_AUTH_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *AnthropicProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AnthropicProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var baseUrl string
	if !data.BaseUrl.IsNull() {
		baseUrl = data.BaseUrl.ValueString()
	} else if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "" {
		baseUrl = v
	} else {
		baseUrl = "https://api.anthropic.com"
	}

	var apiKey string
	if !data.ApiKey.IsNull() {
		apiKey = data.ApiKey.ValueString()
	} else if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		apiKey = v
	}

	if apiKey != "" && !strings.HasPrefix(apiKey, "sk-ant-admin") {
		resp.Diagnostics.AddError("api_key is invalid", "api_key is invalid. It must start with 'sk-ant-admin'.")
		return
	}

	var authToken string
	if !data.AuthToken.IsNull() {
		authToken = data.AuthToken.ValueString()
	} else if v := os.Getenv("ANTHROPIC_AUTH_TOKEN"); v != "" {
		authToken = v
	}

	if authToken != "" && !strings.HasPrefix(authToken, "sk-ant-oat") {
		resp.Diagnostics.AddError("auth_token is invalid", "auth_token is invalid. It must start with 'sk-ant-oat'.")
		return
	}

	if baseUrl == "" {
		resp.Diagnostics.AddError("base_url is required", "base_url is required")
		return
	}

	transport := http.DefaultTransport

	// Logging
	transport = logging.NewLoggingHTTPTransport(transport)

	// Retry
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{Transport: transport}
	retryClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	retryClient.Logger = nil
	retryClient.RetryMax = 10
	transport = retryClient.StandardClient().Transport

	client, err := apiclient.NewClientWithResponses(
		baseUrl,
		apiclient.WithHTTPClient(&http.Client{Transport: transport}),
		apiclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("anthropic-version", "2023-06-01")
			return nil
		}),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to create API client", err.Error())
		return
	}

	providerData := &providerdata.ProviderData{
		ApiKey:    apiKey,
		AuthToken: authToken,
		Client:    client,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *AnthropicProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationInviteResource,
		NewServiceAccountResource,
		NewWorkspaceMemberResource,
		NewWorkspaceResource,
	}
}

func (p *AnthropicProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationDataSource,
		NewOrganizationInvitesDataSource,
		NewUserDataSource,
		NewUsersDataSource,
		NewWorkspaceDataSource,
		NewWorkspaceMemberDataSource,
		NewWorkspaceMembersDataSource,
		NewWorkspacesDataSource,
	}
}

func (p *AnthropicProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AnthropicProvider{
			version: version,
		}
	}
}
