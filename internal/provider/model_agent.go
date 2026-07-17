package provider

import (
	"context"
	"fmt"

	"github.com/frank-bee/terraform-provider-anthropic/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AgentModel struct {
	Id          types.String     `tfsdk:"id"`
	Version     types.String     `tfsdk:"version"`
	Name        types.String     `tfsdk:"name"`
	Description types.String     `tfsdk:"description"`
	System      types.String     `tfsdk:"system"`
	Model       types.String     `tfsdk:"model"`
	Metadata    types.Map        `tfsdk:"metadata"`
	Tools       []AgentToolModel `tfsdk:"tools"`
	McpServers  []McpServerModel `tfsdk:"mcp_servers"`
	Skills      []SkillModel     `tfsdk:"skills"`
}

type AgentToolModel struct {
	Type          types.String                 `tfsdk:"type"`
	DefaultConfig *AgentToolDefaultConfigModel `tfsdk:"default_config"`
	Configs       []AgentToolConfigModel       `tfsdk:"configs"`
}

type AgentToolDefaultConfigModel struct {
	Enabled          types.Bool                      `tfsdk:"enabled"`
	PermissionPolicy *AgentToolPermissionPolicyModel `tfsdk:"permission_policy"`
}

type AgentToolPermissionPolicyModel struct {
	Type types.String `tfsdk:"type"`
}

// AgentToolConfigModel is a per-tool override within a toolset (the allowlist entry).
type AgentToolConfigModel struct {
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

type McpServerModel struct {
	Name          types.String                 `tfsdk:"name"`
	Type          types.String                 `tfsdk:"type"`
	Url           types.String                 `tfsdk:"url"`
	DefaultConfig *AgentToolDefaultConfigModel `tfsdk:"default_config"`
	Configs       []AgentToolConfigModel       `tfsdk:"configs"`
}

type SkillModel struct {
	SkillId types.String `tfsdk:"skill_id"`
	Type    types.String `tfsdk:"type"`
	Version types.String `tfsdk:"version"`
}

func (m *AgentModel) Fill(a apiclient.Agent) error {
	m.Id = types.StringValue(a.Id)
	m.Version = types.StringValue(a.Version)
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringPointerValue(a.Description)
	m.System = types.StringPointerValue(a.System)

	if a.Model != nil {
		m.Model = types.StringValue(a.Model.Id)
	} else {
		m.Model = types.StringNull()
	}

	if a.Metadata != nil && len(*a.Metadata) > 0 {
		elems := make(map[string]string, len(*a.Metadata))
		for k, v := range *a.Metadata {
			elems[k] = v
		}
		mv, diags := types.MapValueFrom(context.Background(), types.StringType, elems)
		if diags.HasError() {
			return fmt.Errorf("failed to build metadata map: %s", diags)
		}
		m.Metadata = mv
	} else {
		m.Metadata = types.MapNull(types.StringType)
	}

	if a.Tools != nil {
		var tools []AgentToolModel
		for _, t := range *a.Tools {
			// Skip mcp_toolset entries — they are auto-generated from mcp_servers
			if t.Type == "mcp_toolset" {
				continue
			}
			tm := AgentToolModel{
				Type:    types.StringValue(t.Type),
				Configs: parseToolConfigs(t.Configs),
			}
			if t.DefaultConfig != nil {
				dc := &AgentToolDefaultConfigModel{
					Enabled: types.BoolPointerValue(t.DefaultConfig.Enabled),
				}
				if t.DefaultConfig.PermissionPolicy != nil {
					dc.PermissionPolicy = &AgentToolPermissionPolicyModel{
						Type: types.StringValue(t.DefaultConfig.PermissionPolicy.Type),
					}
				}
				tm.DefaultConfig = dc
			}
			tools = append(tools, tm)
		}
		if tools == nil {
			tools = []AgentToolModel{}
		}
		m.Tools = tools
	} else {
		m.Tools = []AgentToolModel{}
	}

	// Index mcp_toolset default_configs by server name so we can surface
	// them on the corresponding mcp_servers entry.
	mcpToolsetCfg := map[string]*AgentToolDefaultConfigModel{}
	mcpToolsetConfigs := map[string][]AgentToolConfigModel{}
	if a.Tools != nil {
		for _, t := range *a.Tools {
			if t.Type != "mcp_toolset" || t.McpServerName == nil {
				continue
			}
			if t.DefaultConfig != nil {
				dc := &AgentToolDefaultConfigModel{
					Enabled: types.BoolPointerValue(t.DefaultConfig.Enabled),
				}
				if t.DefaultConfig.PermissionPolicy != nil {
					dc.PermissionPolicy = &AgentToolPermissionPolicyModel{
						Type: types.StringValue(t.DefaultConfig.PermissionPolicy.Type),
					}
				}
				mcpToolsetCfg[*t.McpServerName] = dc
			}
			if cfgs := parseToolConfigs(t.Configs); cfgs != nil {
				mcpToolsetConfigs[*t.McpServerName] = cfgs
			}
		}
	}

	if a.McpServers != nil {
		servers := make([]McpServerModel, len(*a.McpServers))
		for i, s := range *a.McpServers {
			servers[i] = McpServerModel{
				Name:          types.StringValue(s.Name),
				Type:          types.StringValue(s.Type),
				Url:           types.StringValue(s.Url),
				DefaultConfig: mcpToolsetCfg[s.Name],
				Configs:       mcpToolsetConfigs[s.Name],
			}
		}
		m.McpServers = servers
	} else {
		m.McpServers = []McpServerModel{}
	}

	if a.Skills != nil {
		skills := make([]SkillModel, len(*a.Skills))
		for i, s := range *a.Skills {
			skills[i] = SkillModel{
				SkillId: types.StringValue(s.SkillId),
				Type:    types.StringValue(s.Type),
				Version: types.StringValue(s.Version),
			}
		}
		m.Skills = skills
	} else {
		m.Skills = []SkillModel{}
	}

	return nil
}
