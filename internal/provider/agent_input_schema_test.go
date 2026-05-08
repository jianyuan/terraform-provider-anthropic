package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

// Regression test for codex round 5 P3: a custom tool's input_schema is now a
// jsontypes.Normalized so re-marshalling on the API echo doesn't trigger a
// drift against user-written JSON whose key order or whitespace differs from
// Go's canonical form.
//
// We exercise StringSemanticEquals with two equivalent JSON strings whose
// surface form differs.
func TestAgentToolInputSchema_SemanticEquality(t *testing.T) {
	ctx := context.Background()

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
// custom tool input_schema. The previous implementation produced a
// types.String, which would cause perpetual diffs against semantically
// equivalent user input.
func TestAgentModel_Fill_customToolInputSchemaIsNormalized(t *testing.T) {
	ctx := context.Background()

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
