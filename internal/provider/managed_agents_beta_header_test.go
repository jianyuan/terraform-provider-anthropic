package provider

import (
	"net/http"
	"net/url"
	"testing"
)

func TestNewManagedAgentsBetaHeaderEditor(t *testing.T) {
	const want = "managed-agents-2026-04-01"

	cases := []struct {
		name     string
		basePath string
		path     string
		set      bool
	}{
		// Direct api.anthropic.com paths (basePath is empty).
		{"agents direct", "", "/v1/agents", true},
		{"agent get direct", "", "/v1/agents/agent_01ABC", true},
		{"agent archive direct", "", "/v1/agents/agent_01ABC/archive", true},
		{"environments direct", "", "/v1/environments", true},
		{"vault archive direct", "", "/v1/vaults/vlt_01XYZ/archive", true},
		{"organizations users direct", "", "/v1/organizations/users", false},
		{"organizations workspace direct", "", "/v1/organizations/workspaces/wrk_01ABC", false},

		// Proxied base_url with a path component (e.g. "https://proxy/anthropic/").
		// Regression for codex review P1: the old HasPrefix on a hard-coded
		// "/v1/..." string missed these. The factory now parameterises the
		// prefix on basePath so HasPrefix still works.
		{"agents through proxy", "/anthropic", "/anthropic/v1/agents", true},
		{"environment get through gateway", "/api/proxy", "/api/proxy/v1/environments/env_01ABC", true},
		{"vaults under tenant prefix", "/tenants/t1", "/tenants/t1/v1/vaults", true},
		{"organizations workspaces proxied", "/anthropic", "/anthropic/v1/organizations/workspaces", false},

		// basePath supplied with a trailing slash should be normalised.
		{"trailing slash basePath", "/anthropic/", "/anthropic/v1/agents", true},

		// A request whose path doesn't share the basePath prefix must NOT
		// receive the beta header. This guards against substring-style matches
		// erroneously firing on unrelated paths that merely happen to contain
		// "/v1/agents".
		{"foreign path containing v1/agents substring", "/anthropic", "/elsewhere/v1/agents", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			editor := newManagedAgentsBetaHeaderEditor(tc.basePath)
			req := &http.Request{
				URL:    &url.URL{Path: tc.path},
				Header: http.Header{},
			}
			if err := editor(t.Context(), req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := req.Header.Get("anthropic-beta")
			if tc.set {
				if got != want {
					t.Errorf("basePath=%q path=%q: expected anthropic-beta=%q, got %q", tc.basePath, tc.path, want, got)
				}
			} else {
				if got != "" {
					t.Errorf("basePath=%q path=%q: expected no anthropic-beta header, got %q", tc.basePath, tc.path, got)
				}
			}
		})
	}
}
