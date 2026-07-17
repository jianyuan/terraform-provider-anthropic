package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/frank-bee/terraform-provider-anthropic/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mapFromTFMap converts a types.Map of string values into the pointer-to-map
// shape the API client expects. Returns nil when the Terraform value is null
// or unknown.
func mapFromTFMap(ctx context.Context, m types.Map, diags *diag.Diagnostics) *map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	out := map[string]string{}
	d := m.ElementsAs(ctx, &out, false)
	diags.Append(d...)
	if d.HasError() {
		return nil
	}
	return &out
}

// buildStringList converts a types.List of strings into the pointer-to-slice
// shape the API client expects. Returns nil when the Terraform value is null
// or unknown.
func buildStringList(ctx context.Context, l types.List, diags *diag.Diagnostics) *[]string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var out []string
	diags.Append(l.ElementsAs(ctx, &out, false)...)
	if diags.HasError() {
		return nil
	}
	return &out
}

// buildAgentTools converts the Terraform tools + mcp_servers model into the
// flat slice of apiclient.AgentTool entries the API expects. For each MCP
// server an auto-generated `mcp_toolset` entry is appended.
func buildAgentTools(tools []AgentToolModel, servers []McpServerModel) []apiclient.AgentTool {
	var out []apiclient.AgentTool
	for _, t := range tools {
		at := apiclient.AgentTool{Type: t.Type.ValueString()}
		if dc := buildDefaultConfig(t.DefaultConfig); dc != nil {
			at.DefaultConfig = dc
		}
		if cfgs := buildToolConfigs(t.Configs); cfgs != nil {
			at.Configs = cfgs
		}
		out = append(out, at)
	}
	for _, s := range servers {
		at := apiclient.AgentTool{
			Type:          "mcp_toolset",
			McpServerName: ptrTo(s.Name.ValueString()),
		}
		if dc := buildDefaultConfig(s.DefaultConfig); dc != nil {
			at.DefaultConfig = dc
		}
		if cfgs := buildToolConfigs(s.Configs); cfgs != nil {
			at.Configs = cfgs
		}
		out = append(out, at)
	}
	return out
}

// buildToolConfigs converts the Terraform per-tool override list into the
// untyped map slice the API client carries. Returns nil when empty.
func buildToolConfigs(configs []AgentToolConfigModel) *[]map[string]interface{} {
	if len(configs) == 0 {
		return nil
	}
	out := make([]map[string]interface{}, len(configs))
	for i, c := range configs {
		m := map[string]interface{}{"name": c.Name.ValueString()}
		if !c.Enabled.IsNull() && !c.Enabled.IsUnknown() {
			m["enabled"] = c.Enabled.ValueBool()
		}
		out[i] = m
	}
	return &out
}

// parseToolConfigs reads the API's per-tool config maps back into the Terraform
// model. Only `name` and `enabled` are surfaced; any server-normalized fields
// (e.g. an inherited permission_policy) are intentionally ignored so they don't
// produce perpetual plan drift.
func parseToolConfigs(raw *[]map[string]interface{}) []AgentToolConfigModel {
	if raw == nil || len(*raw) == 0 {
		return nil
	}
	out := make([]AgentToolConfigModel, 0, len(*raw))
	for _, m := range *raw {
		c := AgentToolConfigModel{Name: types.StringNull(), Enabled: types.BoolNull()}
		if name, ok := m["name"].(string); ok {
			c.Name = types.StringValue(name)
		}
		if en, ok := m["enabled"].(bool); ok {
			c.Enabled = types.BoolValue(en)
		}
		out = append(out, c)
	}
	return out
}

func buildDefaultConfig(m *AgentToolDefaultConfigModel) *apiclient.AgentToolDefaultConfig {
	if m == nil {
		return nil
	}
	out := &apiclient.AgentToolDefaultConfig{}
	if !m.Enabled.IsNull() && !m.Enabled.IsUnknown() {
		out.Enabled = m.Enabled.ValueBoolPointer()
	}
	if m.PermissionPolicy != nil && !m.PermissionPolicy.Type.IsNull() && !m.PermissionPolicy.Type.IsUnknown() {
		out.PermissionPolicy = &apiclient.AgentToolPermissionPolicy{
			Type: m.PermissionPolicy.Type.ValueString(),
		}
	}
	if out.Enabled == nil && out.PermissionPolicy == nil {
		return nil
	}
	return out
}

// jsonListOrNull converts an API list of free-form objects (e.g. a
// deployment's initial_events/resources) into a types.List of JSON-encoded
// strings, one per element, so the Terraform schema doesn't need to model
// every possible event/resource shape. Empty/nil lists become null so they
// don't show up as drift against an unset optional attribute.
func jsonListOrNull(raw *[]map[string]interface{}) (types.List, error) {
	if raw == nil || len(*raw) == 0 {
		return types.ListNull(types.StringType), nil
	}
	elems := make([]attr.Value, 0, len(*raw))
	for _, m := range *raw {
		b, err := json.Marshal(m)
		if err != nil {
			return types.ListNull(types.StringType), err
		}
		elems = append(elems, types.StringValue(string(b)))
	}
	return types.ListValueMust(types.StringType, elems), nil
}

// buildJSONList is the inverse of jsonListOrNull: it parses each JSON-encoded
// string in a types.List back into a free-form map for the API request body.
// Returns nil when the Terraform value is null or unknown.
func buildJSONList(ctx context.Context, l types.List, diags *diag.Diagnostics) *[]map[string]interface{} {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var raw []string
	diags.Append(l.ElementsAs(ctx, &raw, false)...)
	if diags.HasError() {
		return nil
	}
	out := make([]map[string]interface{}, len(raw))
	for i, s := range raw {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			diags.AddError("Invalid JSON", fmt.Sprintf("Element %d is not a valid JSON object: %s", i, err))
			return nil
		}
		out[i] = m
	}
	return &out
}

func BuildTwoPartId(a, b string) string {
	return fmt.Sprintf("%s/%s", a, b)
}

func ptrTo[T any](v T) *T {
	return &v
}

func SplitTwoPartId(id, a, b string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected %s/%s", id, a, b)
	}
	return parts[0], parts[1], nil
}
