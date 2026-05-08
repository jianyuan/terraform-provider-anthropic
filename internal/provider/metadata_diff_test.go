package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Regression test for codex round 4 P2: when state has no metadata (null) and
// the user explicitly sets `metadata = {}`, the diff is empty (no keys to
// add, none to remove). The previous implementation returned a nil pointer,
// which caused the update body to omit `metadata` entirely — the API would
// then leave the remote value unchanged and Terraform could never converge.
//
// metadataDiff now always returns a non-nil pointer so the caller can send
// `metadata: {}` explicitly.
func TestMetadataDiff(t *testing.T) {
	ctx := t.Context()

	mustMap := func(t *testing.T, m map[string]string) types.Map {
		t.Helper()
		v, d := types.MapValueFrom(ctx, types.StringType, m)
		if d.HasError() {
			t.Fatalf("MapValueFrom: %v", d)
		}
		return v
	}

	cases := []struct {
		name  string
		state types.Map
		plan  types.Map
		// expectNil documents that metadataDiff returns nil only on error.
		expectNil  bool
		expectKeys map[string]string
	}{
		{
			name:       "add new key from null state",
			state:      types.MapNull(types.StringType),
			plan:       mustMap(t, map[string]string{"env": "dev"}),
			expectKeys: map[string]string{"env": "dev"},
		},
		{
			name:       "remove key sends empty-string tombstone",
			state:      mustMap(t, map[string]string{"env": "dev"}),
			plan:       types.MapNull(types.StringType),
			expectKeys: map[string]string{"env": ""},
		},
		{
			name:       "update existing key",
			state:      mustMap(t, map[string]string{"env": "dev"}),
			plan:       mustMap(t, map[string]string{"env": "prod"}),
			expectKeys: map[string]string{"env": "prod"},
		},
		{
			name:       "mix add and delete",
			state:      mustMap(t, map[string]string{"env": "dev", "team": "platform"}),
			plan:       mustMap(t, map[string]string{"env": "prod"}),
			expectKeys: map[string]string{"env": "prod", "team": ""},
		},
		{
			// Regression: plan = {} and state = null. Previously the diff was
			// nil and metadata was omitted from the request, so an "explicit
			// empty map" config could never converge.
			name:       "empty plan with null state still returns payload",
			state:      types.MapNull(types.StringType),
			plan:       mustMap(t, map[string]string{}),
			expectKeys: map[string]string{},
		},
		{
			name:       "empty plan with non-empty state deletes all keys",
			state:      mustMap(t, map[string]string{"a": "1", "b": "2"}),
			plan:       mustMap(t, map[string]string{}),
			expectKeys: map[string]string{"a": "", "b": ""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, diags := metadataDiff(ctx, tc.state, tc.plan)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got == nil {
				t.Fatalf("expected non-nil payload, got nil")
			}
			if len(*got) != len(tc.expectKeys) {
				t.Fatalf("expected %d keys, got %d (%v)", len(tc.expectKeys), len(*got), *got)
			}
			for k, v := range tc.expectKeys {
				if (*got)[k] != v {
					t.Errorf("key %q: expected %q, got %q", k, v, (*got)[k])
				}
			}
		})
	}
}
