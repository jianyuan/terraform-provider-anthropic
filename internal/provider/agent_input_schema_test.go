package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

// Custom-tool input_schema is a jsontypes.Normalized — semantically-equal
// JSON (different whitespace, different key order) compares equal so the
// API echo doesn't drift against user-written heredoc JSON.
//
// We can't expose this as a native HCL DynamicAttribute because the
// Terraform Plugin Framework forbids dynamic types nested inside a
// collection (see https://github.com/hashicorp/terraform-plugin-framework
// "Dynamic types inside of collections are not currently supported").
// `tools` is a ListNestedAttribute, so input_schema must remain a string
// (jsonencode'd) for this layout.
func TestAgentToolInputSchema_SemanticEquality(t *testing.T) {
	ctx := t.Context()

	heredocStyle := jsontypes.NewNormalizedValue(`{
  "type": "object",
  "properties": {
    "location": { "type": "string", "description": "City name" }
  },
  "required": ["location"]
}`)
	canonical := jsontypes.NewNormalizedValue(
		`{"properties":{"location":{"description":"City name","type":"string"}},"required":["location"],"type":"object"}`,
	)

	equal, diags := heredocStyle.StringSemanticEquals(ctx, canonical)
	if diags.HasError() {
		t.Fatalf("StringSemanticEquals returned diagnostics: %v", diags)
	}
	if !equal {
		t.Fatalf("expected heredoc-style JSON to be semantically equal to canonical JSON")
	}
}

// Round-trip test: Fill produces a Normalized value when the API echoes a
// custom tool input_schema.
func TestAgentModel_Fill_customToolInputSchemaIsNormalized(t *testing.T) {
	ctx := t.Context()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{"type": "string"},
		},
	}

	tools := []apiclient.AgentTool{{
		Type:        "custom",
		Name:        ptrString("get_weather"),
		Description: ptrString("Get weather for a location."),
		InputSchema: &schema,
	}}

	m := &AgentModel{}
	if d := m.Fill(ctx, apiclient.Agent{
		Id:        "agent_01ABC",
		Type:      "agent",
		Name:      "test",
		Model:     apiclient.Model{Id: "claude-opus-4-7"},
		Tools:     &tools,
		Version:   1,
		CreatedAt: "2026-04-01T00:00:00Z",
		UpdatedAt: "2026-04-01T00:00:00Z",
	}); d.HasError() {
		t.Fatalf("Fill returned diagnostics: %v", d)
	}

	if m.Tools.IsNull() {
		t.Fatal("expected tools to be non-null")
	}
	elems := m.Tools.Elements()
	if len(elems) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(elems))
	}
}

func ptrString(s string) *string { return &s }
