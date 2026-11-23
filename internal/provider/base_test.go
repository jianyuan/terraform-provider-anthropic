package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/providerdata"
)

// Every resource and data source must set exactly one of requiresApiKey or requiresAuthToken.
func TestCredentialRequirements(t *testing.T) {
	ctx := context.Background()
	p := &AnthropicProvider{version: "test"}
	pd := &providerdata.ProviderData{ApiKey: "sk-ant-admin-test", AuthToken: "sk-ant-oat-test"}

	for _, f := range p.Resources(ctx) {
		r, ok := f().(resource.ResourceWithConfigure)
		if !ok {
			t.Fatalf("%T does not implement ResourceWithConfigure", f())
		}
		var resp resource.ConfigureResponse
		r.Configure(ctx, resource.ConfigureRequest{ProviderData: pd}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("%T: %v", r, resp.Diagnostics)
		}
	}
	for _, f := range p.DataSources(ctx) {
		d, ok := f().(datasource.DataSourceWithConfigure)
		if !ok {
			t.Fatalf("%T does not implement DataSourceWithConfigure", f())
		}
		var resp datasource.ConfigureResponse
		d.Configure(ctx, datasource.ConfigureRequest{ProviderData: pd}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("%T: %v", d, resp.Diagnostics)
		}
	}
}
