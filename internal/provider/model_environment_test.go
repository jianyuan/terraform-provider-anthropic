package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

// Regression test for codex review round 2 P2: when the API omits
// allow_mcp_servers / allow_package_managers from the networking response
// (their schema marks them optional), Fill must coerce them to false (the
// documented default) rather than leaving them as null. With config marked
// RequiresReplace, a null-vs-false drift would cause Terraform to replan a
// replacement on every refresh.
func TestEnvironmentModel_Fill_networkingFlagsDefaultFalse(t *testing.T) {
	ctx := t.Context()

	m := &EnvironmentModel{}
	diags := m.Fill(ctx, apiclient.Environment{
		Id:        "env_01ABC",
		Type:      "environment",
		Name:      "test",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
		Config: apiclient.EnvironmentConfig{
			Type: "cloud",
			Networking: apiclient.Networking{
				Type: "limited",
				// AllowMcpServers and AllowPackageManagers intentionally nil.
				AllowedHosts: &[]string{"api.example.com"},
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", diags)
	}

	cfg, ok := m.Config.Attributes()["networking"].(types.Object)
	if !ok {
		t.Fatalf("expected networking attribute to be types.Object, got %T", m.Config.Attributes()["networking"])
	}

	var net environmentNetworkingModel
	if d := cfg.As(ctx, &net, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding networking: %v", d)
	}

	for _, tc := range []struct {
		name string
		v    types.Bool
	}{
		{"allow_mcp_servers", net.AllowMcpServers},
		{"allow_package_managers", net.AllowPackageManagers},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.v.IsNull() {
				t.Errorf("%s: expected false (default), got null", tc.name)
			}
			if tc.v.ValueBool() != false {
				t.Errorf("%s: expected false, got %v", tc.name, tc.v.ValueBool())
			}
		})
	}
}

func TestEnvironmentModel_Fill_networkingFlagsExplicitTrue(t *testing.T) {
	ctx := t.Context()

	tr := true
	m := &EnvironmentModel{}
	diags := m.Fill(ctx, apiclient.Environment{
		Id:        "env_01ABC",
		Type:      "environment",
		Name:      "test",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
		Config: apiclient.EnvironmentConfig{
			Type: "cloud",
			Networking: apiclient.Networking{
				Type:                 "limited",
				AllowMcpServers:      &tr,
				AllowPackageManagers: &tr,
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", diags)
	}

	cfg := m.Config.Attributes()["networking"].(types.Object)
	var net environmentNetworkingModel
	if d := cfg.As(ctx, &net, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding networking: %v", d)
	}
	if !net.AllowMcpServers.ValueBool() {
		t.Error("allow_mcp_servers: expected true")
	}
	if !net.AllowPackageManagers.ValueBool() {
		t.Error("allow_package_managers: expected true")
	}
}

// Regression test for codex round 9 P1: when the API echoes
// `packages: {}` (object exists but every per-manager list is nil) for an
// environment where the user did NOT configure packages at all, Fill must
// collapse the response back to ObjectNull. Otherwise state would have a
// phantom `{ apt = null, ... }` object that mismatches a config that omits
// the `packages` attribute, and `config` (RequiresReplace) would force a
// replacement on every refresh.
func TestEnvironmentModel_Fill_packagesEmptyObjectMirrorsNullLists(t *testing.T) {
	ctx := t.Context()

	m := &EnvironmentModel{}
	diags := m.Fill(ctx, apiclient.Environment{
		Id:        "env_01ABC",
		Type:      "environment",
		Name:      "test",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
		Config: apiclient.EnvironmentConfig{
			Type: "cloud",
			Networking: apiclient.Networking{
				Type: "unrestricted",
			},
			Packages: &apiclient.Packages{},
		},
	})
	if diags.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", diags)
	}

	pkgs := m.Config.Attributes()["packages"].(types.Object)
	if !pkgs.IsNull() {
		t.Errorf("expected packages to collapse to ObjectNull when API echoed packages: {} (codex round 9 fix), got non-null with attributes %v", pkgs.Attributes())
	}
}

// When at least one per-manager list is set, the packages object stays
// populated.
func TestEnvironmentModel_Fill_packagesWithOneManagerStaysPopulated(t *testing.T) {
	ctx := t.Context()

	m := &EnvironmentModel{}
	diags := m.Fill(ctx, apiclient.Environment{
		Id:        "env_01ABC",
		Type:      "environment",
		Name:      "test",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
		Config: apiclient.EnvironmentConfig{
			Type: "cloud",
			Networking: apiclient.Networking{
				Type: "unrestricted",
			},
			Packages: &apiclient.Packages{
				Pip: &[]string{"pandas"},
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", diags)
	}

	pkgs := m.Config.Attributes()["packages"].(types.Object)
	if pkgs.IsNull() {
		t.Fatal("expected packages to be non-null when at least one manager list is set")
	}
	pip := pkgs.Attributes()["pip"].(types.List)
	if pip.IsNull() {
		t.Fatal("expected pip to be a populated list")
	}
	if got := len(pip.Elements()); got != 1 {
		t.Errorf("expected 1 pip element, got %d", got)
	}
}

// When the API does not echo any packages object at all (e.g. user didn't
// configure packages), Fill must keep the parent attribute null — otherwise
// users with no packages would suddenly see a populated empty object in
// state.
func TestEnvironmentModel_Fill_packagesAbsentStaysNull(t *testing.T) {
	ctx := t.Context()

	m := &EnvironmentModel{}
	diags := m.Fill(ctx, apiclient.Environment{
		Id:        "env_01ABC",
		Type:      "environment",
		Name:      "test",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
		Config: apiclient.EnvironmentConfig{
			Type: "cloud",
			Networking: apiclient.Networking{
				Type: "unrestricted",
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", diags)
	}

	pkgs := m.Config.Attributes()["packages"].(types.Object)
	if !pkgs.IsNull() {
		t.Errorf("expected packages to be null when API omits it, got non-null")
	}
}

// Regression test for codex round 7 P1: declaring the boolean networking
// flags Computed-only made them sticky — a user who once set
// allow_mcp_servers/allow_package_managers to true could not later
// re-tighten the environment by removing the line. The current schema uses
// Default(false) instead, which means an absent attribute always plans as
// `false` regardless of prior state. We verify the model layer here; the
// schema-level Default behaviour is exercised by the framework itself.
func TestEnvironmentModel_Fill_explicitFalseRoundTrips(t *testing.T) {
	ctx := t.Context()

	f := false
	m := &EnvironmentModel{}
	diags := m.Fill(ctx, apiclient.Environment{
		Id:        "env_01ABC",
		Type:      "environment",
		Name:      "test",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
		Config: apiclient.EnvironmentConfig{
			Type: "cloud",
			Networking: apiclient.Networking{
				Type:                 "limited",
				AllowedHosts:         &[]string{"api.example.com"},
				AllowMcpServers:      &f,
				AllowPackageManagers: &f,
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", diags)
	}

	cfg := m.Config.Attributes()["networking"].(types.Object)
	var net environmentNetworkingModel
	if d := cfg.As(ctx, &net, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding networking: %v", d)
	}
	if net.AllowMcpServers.IsNull() || net.AllowMcpServers.ValueBool() {
		t.Errorf("allow_mcp_servers: expected explicit false, got %v (null=%v)", net.AllowMcpServers.ValueBool(), net.AllowMcpServers.IsNull())
	}
	if net.AllowPackageManagers.IsNull() || net.AllowPackageManagers.ValueBool() {
		t.Errorf("allow_package_managers: expected explicit false, got %v (null=%v)", net.AllowPackageManagers.ValueBool(), net.AllowPackageManagers.IsNull())
	}
}

// allowed_hosts is Optional-only by design: removing it from configuration
// should tighten the policy on next apply. Fill must mirror the API echo
// (nil → null) rather than coercing to []. Drift would only occur if the
// API normalises an explicit `[]` to nil; the Anthropic SDK pattern is to
// echo back the same shape, so we accept that limitation as-is.
func TestEnvironmentModel_Fill_allowedHostsOmittedStaysNull(t *testing.T) {
	ctx := t.Context()

	m := &EnvironmentModel{}
	diags := m.Fill(ctx, apiclient.Environment{
		Id:        "env_01ABC",
		Type:      "environment",
		Name:      "test",
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
		Config: apiclient.EnvironmentConfig{
			Type: "cloud",
			Networking: apiclient.Networking{
				Type: "limited",
				// AllowedHosts intentionally nil — the API may legally omit
				// it from the response when the underlying value was [].
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", diags)
	}

	cfg := m.Config.Attributes()["networking"].(types.Object)
	var net environmentNetworkingModel
	if d := cfg.As(ctx, &net, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("decoding networking: %v", d)
	}
	if net.AllowedHosts.IsNull() {
		// Expected: schema is Optional-only, Fill mirrors API → null.
		return
	}
	t.Fatalf("expected allowed_hosts to be ListNull (Optional-only schema, mirrored from nil API value), got non-null with %d elements", len(net.AllowedHosts.Elements()))
}
