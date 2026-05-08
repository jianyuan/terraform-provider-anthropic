package provider

import (
	"context"
	"testing"

	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

// The API treats `metadata = {}` and an absent metadata field identically.
// To produce convergent state, Fill collapses both nil and empty maps to
// MapNull; the matching schema validator (mapvalidator.SizeAtLeast(1)) keeps
// the user from typing `metadata = {}` in config in the first place.

func TestVaultModel_Fill_metadataNilStaysNull(t *testing.T) {
	ctx := context.Background()

	m := &VaultModel{}
	if d := m.Fill(ctx, apiclient.Vault{
		Id:          "vlt_01ABC",
		Type:        "vault",
		DisplayName: "Alice",
		CreatedAt:   "2026-04-01T00:00:00Z",
		UpdatedAt:   "2026-04-01T00:00:00Z",
		Metadata:    nil,
	}); d.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", d)
	}

	if !m.Metadata.IsNull() {
		t.Fatal("expected Metadata to be MapNull when API returns nil, got non-null")
	}
}

func TestVaultModel_Fill_metadataEmptyMapCollapsedToNull(t *testing.T) {
	ctx := context.Background()
	empty := map[string]string{}

	m := &VaultModel{}
	if d := m.Fill(ctx, apiclient.Vault{
		Id:          "vlt_01ABC",
		Type:        "vault",
		DisplayName: "Alice",
		CreatedAt:   "2026-04-01T00:00:00Z",
		UpdatedAt:   "2026-04-01T00:00:00Z",
		Metadata:    &empty,
	}); d.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", d)
	}

	if !m.Metadata.IsNull() {
		t.Fatal("expected empty API metadata map to collapse to MapNull (round 5 fix), got non-null")
	}
}

func TestVaultModel_Fill_metadataPopulated(t *testing.T) {
	ctx := context.Background()
	m := &VaultModel{}
	md := map[string]string{"env": "dev"}
	if d := m.Fill(ctx, apiclient.Vault{
		Id:          "vlt_01ABC",
		Type:        "vault",
		DisplayName: "Alice",
		CreatedAt:   "2026-04-01T00:00:00Z",
		UpdatedAt:   "2026-04-01T00:00:00Z",
		Metadata:    &md,
	}); d.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", d)
	}

	if m.Metadata.IsNull() {
		t.Fatal("expected populated metadata to be non-null")
	}
	if got := len(m.Metadata.Elements()); got != 1 {
		t.Fatalf("expected 1 metadata element, got %d", got)
	}
}

func TestAgentModel_Fill_metadataNilStaysNull(t *testing.T) {
	ctx := context.Background()

	m := &AgentModel{}
	if d := m.Fill(ctx, apiclient.Agent{
		Id:        "agent_01ABC",
		Type:      "agent",
		Name:      "test",
		Model:     apiclient.Model{Id: "claude-opus-4-7"},
		Metadata:  nil,
		Version:   1,
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	}); d.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", d)
	}

	if !m.Metadata.IsNull() {
		t.Fatal("expected Metadata to be MapNull when API returns nil, got non-null")
	}
}

func TestAgentModel_Fill_metadataEmptyMapCollapsedToNull(t *testing.T) {
	ctx := context.Background()
	empty := map[string]string{}

	m := &AgentModel{}
	if d := m.Fill(ctx, apiclient.Agent{
		Id:        "agent_01ABC",
		Type:      "agent",
		Name:      "test",
		Model:     apiclient.Model{Id: "claude-opus-4-7"},
		Metadata:  &empty,
		Version:   1,
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	}); d.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", d)
	}

	if !m.Metadata.IsNull() {
		t.Fatal("expected empty API metadata map to collapse to MapNull (round 5 fix), got non-null")
	}
}

func TestAgentModel_Fill_metadataPopulated(t *testing.T) {
	ctx := context.Background()
	m := &AgentModel{}
	md := map[string]string{"env": "dev"}
	if d := m.Fill(ctx, apiclient.Agent{
		Id:        "agent_01ABC",
		Type:      "agent",
		Name:      "test",
		Model:     apiclient.Model{Id: "claude-opus-4-7"},
		Metadata:  &md,
		Version:   1,
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	}); d.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", d)
	}
	if m.Metadata.IsNull() {
		t.Fatal("expected populated metadata to be non-null")
	}
}
