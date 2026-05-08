package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type AgentModel struct {
	Id          types.String `tfsdk:"id"`
	Version     types.Int64  `tfsdk:"version"`
	Name        types.String `tfsdk:"name"`
	Model       types.Object `tfsdk:"model"`
	System      types.String `tfsdk:"system"`
	Description types.String `tfsdk:"description"`
	Tools       types.List   `tfsdk:"tools"`
	McpServers  types.List   `tfsdk:"mcp_servers"`
	Skills      types.List   `tfsdk:"skills"`
	Multiagent  types.Object `tfsdk:"multiagent"`
	Metadata    types.Map    `tfsdk:"metadata"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	ArchivedAt  types.String `tfsdk:"archived_at"`
}

type agentModelModel struct {
	Id      types.String `tfsdk:"id"`
	Speed   types.String `tfsdk:"speed"`
	Version types.String `tfsdk:"version"`
}

func agentModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":      types.StringType,
		"speed":   types.StringType,
		"version": types.StringType,
	}
}

type permissionPolicyModel struct {
	Type types.String `tfsdk:"type"`
}

func permissionPolicyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
	}
}

type toolConfigModel struct {
	Name             types.String `tfsdk:"name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	PermissionPolicy types.Object `tfsdk:"permission_policy"`
}

func toolConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":              types.StringType,
		"enabled":           types.BoolType,
		"permission_policy": types.ObjectType{AttrTypes: permissionPolicyAttrTypes()},
	}
}

type agentToolModel struct {
	Type             types.String         `tfsdk:"type"`
	DefaultConfig    types.Object         `tfsdk:"default_config"`
	Configs          types.List           `tfsdk:"configs"`
	McpServerName    types.String         `tfsdk:"mcp_server_name"`
	Name             types.String         `tfsdk:"name"`
	Description      types.String         `tfsdk:"description"`
	InputSchema      jsontypes.Normalized `tfsdk:"input_schema"`
	PermissionPolicy types.Object         `tfsdk:"permission_policy"`
}

func agentToolAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":              types.StringType,
		"default_config":    types.ObjectType{AttrTypes: toolConfigAttrTypes()},
		"configs":           types.ListType{ElemType: types.ObjectType{AttrTypes: toolConfigAttrTypes()}},
		"mcp_server_name":   types.StringType,
		"name":              types.StringType,
		"description":       types.StringType,
		"input_schema":      jsontypes.NormalizedType{},
		"permission_policy": types.ObjectType{AttrTypes: permissionPolicyAttrTypes()},
	}
}

type mcpServerModel struct {
	Type types.String `tfsdk:"type"`
	Name types.String `tfsdk:"name"`
	Url  types.String `tfsdk:"url"`
}

func mcpServerAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"name": types.StringType,
		"url":  types.StringType,
	}
}

type skillModel struct {
	Type    types.String `tfsdk:"type"`
	SkillId types.String `tfsdk:"skill_id"`
	Version types.String `tfsdk:"version"`
}

func skillAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":     types.StringType,
		"skill_id": types.StringType,
		"version":  types.StringType,
	}
}

type rosterEntryModel struct {
	Type    types.String `tfsdk:"type"`
	Id      types.String `tfsdk:"id"`
	Version types.Int64  `tfsdk:"version"`
}

func rosterEntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":    types.StringType,
		"id":      types.StringType,
		"version": types.Int64Type,
	}
}

type multiagentModel struct {
	Type   types.String `tfsdk:"type"`
	Agents types.List   `tfsdk:"agents"`
}

func multiagentAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":   types.StringType,
		"agents": types.ListType{ElemType: types.ObjectType{AttrTypes: rosterEntryAttrTypes()}},
	}
}

func toolConfigsListNullValue() types.List {
	return types.ListNull(types.ObjectType{AttrTypes: toolConfigAttrTypes()})
}

// Fill populates the model from an API Agent response.
func (m *AgentModel) Fill(ctx context.Context, a apiclient.Agent) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Id = types.StringValue(a.Id)
	m.Version = types.Int64Value(int64(a.Version))
	m.Name = types.StringValue(a.Name)
	m.System = types.StringPointerValue(a.System)
	m.Description = types.StringPointerValue(a.Description)
	m.CreatedAt = types.StringValue(a.CreatedAt)
	m.UpdatedAt = types.StringValue(a.UpdatedAt)
	m.ArchivedAt = types.StringPointerValue(a.ArchivedAt)

	modelObj, d := types.ObjectValueFrom(ctx, agentModelAttrTypes(), agentModelModel{
		Id:      types.StringValue(a.Model.Id),
		Speed:   types.StringPointerValue(a.Model.Speed),
		Version: types.StringPointerValue(a.Model.Version),
	})
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Model = modelObj

	// See comment in model_vault.go: API treats omitted and {} identically,
	// so collapse both to null. The schema validator forbids `metadata = {}`.
	if a.Metadata == nil || len(*a.Metadata) == 0 {
		m.Metadata = types.MapNull(types.StringType)
	} else {
		md, d := types.MapValueFrom(ctx, types.StringType, *a.Metadata)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.Metadata = md
	}

	tools, d := agentToolsToList(ctx, a.Tools)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Tools = tools

	servers, d := mcpServersToList(ctx, a.McpServers)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.McpServers = servers

	skills, d := skillsToList(ctx, a.Skills)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Skills = skills

	ma, d := multiagentToObject(ctx, a.Multiagent)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Multiagent = ma

	return diags
}

func permissionPolicyToObject(ctx context.Context, p *apiclient.PermissionPolicy) (types.Object, diag.Diagnostics) {
	if p == nil {
		return types.ObjectNull(permissionPolicyAttrTypes()), nil
	}
	return types.ObjectValueFrom(ctx, permissionPolicyAttrTypes(), permissionPolicyModel{
		Type: types.StringValue(p.Type),
	})
}

func toolConfigToObject(ctx context.Context, t apiclient.ToolConfig) (types.Object, diag.Diagnostics) {
	pp, d := permissionPolicyToObject(ctx, t.PermissionPolicy)
	if d.HasError() {
		return types.ObjectNull(toolConfigAttrTypes()), d
	}
	return types.ObjectValueFrom(ctx, toolConfigAttrTypes(), toolConfigModel{
		Name:             types.StringPointerValue(t.Name),
		Enabled:          types.BoolPointerValue(t.Enabled),
		PermissionPolicy: pp,
	})
}

func toolConfigsToList(ctx context.Context, configs *[]apiclient.ToolConfig) (types.List, diag.Diagnostics) {
	if configs == nil {
		return toolConfigsListNullValue(), nil
	}
	elems := make([]attr.Value, 0, len(*configs))
	for _, c := range *configs {
		o, d := toolConfigToObject(ctx, c)
		if d.HasError() {
			return toolConfigsListNullValue(), d
		}
		elems = append(elems, o)
	}
	return types.ListValue(types.ObjectType{AttrTypes: toolConfigAttrTypes()}, elems)
}

func agentToolsToList(ctx context.Context, tools *[]apiclient.AgentTool) (types.List, diag.Diagnostics) {
	listType := types.ObjectType{AttrTypes: agentToolAttrTypes()}
	if tools == nil {
		return types.ListNull(listType), nil
	}

	var diags diag.Diagnostics
	elems := make([]attr.Value, 0, len(*tools))
	for _, t := range *tools {
		dc := types.ObjectNull(toolConfigAttrTypes())
		if t.DefaultConfig != nil {
			o, d := toolConfigToObject(ctx, *t.DefaultConfig)
			diags.Append(d...)
			if diags.HasError() {
				return types.ListNull(listType), diags
			}
			dc = o
		}

		configsList, d := toolConfigsToList(ctx, t.Configs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(listType), diags
		}

		pp, d := permissionPolicyToObject(ctx, t.PermissionPolicy)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(listType), diags
		}

		inputSchema := jsontypes.NewNormalizedNull()
		if t.InputSchema != nil {
			b, err := json.Marshal(*t.InputSchema)
			if err != nil {
				diags.AddError("Encoding error", fmt.Sprintf("Failed to marshal tool input_schema: %s", err))
				return types.ListNull(listType), diags
			}
			inputSchema = jsontypes.NewNormalizedValue(string(b))
		}

		obj, d := types.ObjectValueFrom(ctx, agentToolAttrTypes(), agentToolModel{
			Type:             types.StringValue(t.Type),
			DefaultConfig:    dc,
			Configs:          configsList,
			McpServerName:    types.StringPointerValue(t.McpServerName),
			Name:             types.StringPointerValue(t.Name),
			Description:      types.StringPointerValue(t.Description),
			InputSchema:      inputSchema,
			PermissionPolicy: pp,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(listType), diags
		}
		elems = append(elems, obj)
	}
	return types.ListValue(listType, elems)
}

func mcpServersToList(ctx context.Context, servers *[]apiclient.MCPServer) (types.List, diag.Diagnostics) {
	listType := types.ObjectType{AttrTypes: mcpServerAttrTypes()}
	// API echoes an empty list when no MCP servers are configured; treat
	// that as null in state so a config that omits `mcp_servers` doesn't
	// trip "Provider produced inconsistent result after apply".
	if servers == nil || len(*servers) == 0 {
		return types.ListNull(listType), nil
	}
	var diags diag.Diagnostics
	elems := make([]attr.Value, 0, len(*servers))
	for _, s := range *servers {
		obj, d := types.ObjectValueFrom(ctx, mcpServerAttrTypes(), mcpServerModel{
			Type: types.StringValue(s.Type),
			Name: types.StringValue(s.Name),
			Url:  types.StringPointerValue(s.Url),
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(listType), diags
		}
		elems = append(elems, obj)
	}
	return types.ListValue(listType, elems)
}

func skillsToList(ctx context.Context, skills *[]apiclient.Skill) (types.List, diag.Diagnostics) {
	listType := types.ObjectType{AttrTypes: skillAttrTypes()}
	// Same null-vs-empty treatment as mcpServersToList.
	if skills == nil || len(*skills) == 0 {
		return types.ListNull(listType), nil
	}
	var diags diag.Diagnostics
	elems := make([]attr.Value, 0, len(*skills))
	for _, s := range *skills {
		obj, d := types.ObjectValueFrom(ctx, skillAttrTypes(), skillModel{
			Type:    types.StringValue(s.Type),
			SkillId: types.StringValue(s.SkillId),
			Version: types.StringPointerValue(s.Version),
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(listType), diags
		}
		elems = append(elems, obj)
	}
	return types.ListValue(listType, elems)
}

func multiagentToObject(ctx context.Context, m *apiclient.Multiagent) (types.Object, diag.Diagnostics) {
	if m == nil {
		return types.ObjectNull(multiagentAttrTypes()), nil
	}
	var diags diag.Diagnostics
	rosterListType := types.ObjectType{AttrTypes: rosterEntryAttrTypes()}
	elems := make([]attr.Value, 0, len(m.Agents))
	for _, e := range m.Agents {
		var verVal types.Int64
		if e.Version != nil {
			verVal = types.Int64Value(int64(*e.Version))
		} else {
			verVal = types.Int64Null()
		}
		obj, d := types.ObjectValueFrom(ctx, rosterEntryAttrTypes(), rosterEntryModel{
			Type:    types.StringValue(e.Type),
			Id:      types.StringPointerValue(e.Id),
			Version: verVal,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(multiagentAttrTypes()), diags
		}
		elems = append(elems, obj)
	}
	agentsList, d := types.ListValue(rosterListType, elems)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(multiagentAttrTypes()), diags
	}
	obj, d := types.ObjectValueFrom(ctx, multiagentAttrTypes(), multiagentModel{
		Type:   types.StringValue(m.Type),
		Agents: agentsList,
	})
	diags.Append(d...)
	return obj, diags
}

// === Conversion: Terraform model -> API request body ===

func agentModelToAPI(ctx context.Context, modelObj types.Object) (apiclient.Model, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out apiclient.Model
	if modelObj.IsNull() || modelObj.IsUnknown() {
		diags.AddError("Missing model", "model is required")
		return out, diags
	}
	var m agentModelModel
	diags.Append(modelObj.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return out, diags
	}
	out.Id = m.Id.ValueString()
	if !m.Speed.IsNull() && !m.Speed.IsUnknown() {
		v := m.Speed.ValueString()
		out.Speed = &v
	}
	if !m.Version.IsNull() && !m.Version.IsUnknown() {
		v := m.Version.ValueString()
		out.Version = &v
	}
	return out, diags
}

func permissionPolicyFromObject(ctx context.Context, obj types.Object) (*apiclient.PermissionPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	var p permissionPolicyModel
	diags.Append(obj.As(ctx, &p, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	if p.Type.IsNull() || p.Type.IsUnknown() {
		return nil, diags
	}
	return &apiclient.PermissionPolicy{Type: p.Type.ValueString()}, diags
}

func toolConfigFromObject(ctx context.Context, obj types.Object) (apiclient.ToolConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out apiclient.ToolConfig
	if obj.IsNull() || obj.IsUnknown() {
		return out, diags
	}
	var c toolConfigModel
	diags.Append(obj.As(ctx, &c, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return out, diags
	}
	if !c.Name.IsNull() && !c.Name.IsUnknown() {
		v := c.Name.ValueString()
		out.Name = &v
	}
	if !c.Enabled.IsNull() && !c.Enabled.IsUnknown() {
		v := c.Enabled.ValueBool()
		out.Enabled = &v
	}
	pp, d := permissionPolicyFromObject(ctx, c.PermissionPolicy)
	diags.Append(d...)
	out.PermissionPolicy = pp
	return out, diags
}

func toolConfigPtrFromObject(ctx context.Context, obj types.Object) (*apiclient.ToolConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	c, d := toolConfigFromObject(ctx, obj)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return &c, diags
}

func toolConfigsFromList(ctx context.Context, l types.List) (*[]apiclient.ToolConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if l.IsNull() || l.IsUnknown() {
		return nil, diags
	}
	var elems []types.Object
	diags.Append(l.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]apiclient.ToolConfig, 0, len(elems))
	for _, e := range elems {
		c, d := toolConfigFromObject(ctx, e)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out = append(out, c)
	}
	return &out, diags
}

func agentToolsFromList(ctx context.Context, l types.List) (*[]apiclient.AgentTool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if l.IsNull() || l.IsUnknown() {
		return nil, diags
	}
	var elems []types.Object
	diags.Append(l.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]apiclient.AgentTool, 0, len(elems))
	for _, e := range elems {
		var t agentToolModel
		diags.Append(e.As(ctx, &t, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		api := apiclient.AgentTool{Type: t.Type.ValueString()}

		dc, d := toolConfigPtrFromObject(ctx, t.DefaultConfig)
		diags.Append(d...)
		api.DefaultConfig = dc

		configs, d := toolConfigsFromList(ctx, t.Configs)
		diags.Append(d...)
		api.Configs = configs

		if !t.McpServerName.IsNull() && !t.McpServerName.IsUnknown() {
			v := t.McpServerName.ValueString()
			api.McpServerName = &v
		}
		if !t.Name.IsNull() && !t.Name.IsUnknown() {
			v := t.Name.ValueString()
			api.Name = &v
		}
		if !t.Description.IsNull() && !t.Description.IsUnknown() {
			v := t.Description.ValueString()
			api.Description = &v
		}
		if !t.InputSchema.IsNull() && !t.InputSchema.IsUnknown() && t.InputSchema.ValueString() != "" {
			var raw map[string]interface{}
			if err := json.Unmarshal([]byte(t.InputSchema.ValueString()), &raw); err != nil {
				diags.AddError("Invalid input_schema",
					fmt.Sprintf("Failed to parse tool input_schema as JSON object: %s", err))
				return nil, diags
			}
			api.InputSchema = &raw
		}
		pp, d := permissionPolicyFromObject(ctx, t.PermissionPolicy)
		diags.Append(d...)
		api.PermissionPolicy = pp

		if diags.HasError() {
			return nil, diags
		}
		out = append(out, api)
	}
	return &out, diags
}

func mcpServersFromList(ctx context.Context, l types.List) (*[]apiclient.MCPServer, diag.Diagnostics) {
	var diags diag.Diagnostics
	if l.IsNull() || l.IsUnknown() {
		return nil, diags
	}
	var elems []types.Object
	diags.Append(l.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]apiclient.MCPServer, 0, len(elems))
	for _, e := range elems {
		var s mcpServerModel
		diags.Append(e.As(ctx, &s, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		api := apiclient.MCPServer{
			Type: s.Type.ValueString(),
			Name: s.Name.ValueString(),
		}
		if !s.Url.IsNull() && !s.Url.IsUnknown() {
			v := s.Url.ValueString()
			api.Url = &v
		}
		out = append(out, api)
	}
	return &out, diags
}

func skillsFromList(ctx context.Context, l types.List) (*[]apiclient.Skill, diag.Diagnostics) {
	var diags diag.Diagnostics
	if l.IsNull() || l.IsUnknown() {
		return nil, diags
	}
	var elems []types.Object
	diags.Append(l.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]apiclient.Skill, 0, len(elems))
	for _, e := range elems {
		var s skillModel
		diags.Append(e.As(ctx, &s, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		api := apiclient.Skill{
			Type:    s.Type.ValueString(),
			SkillId: s.SkillId.ValueString(),
		}
		if !s.Version.IsNull() && !s.Version.IsUnknown() {
			v := s.Version.ValueString()
			api.Version = &v
		}
		out = append(out, api)
	}
	return &out, diags
}

func multiagentFromObject(ctx context.Context, obj types.Object) (*apiclient.Multiagent, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	var m multiagentModel
	diags.Append(obj.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	out := apiclient.Multiagent{Type: m.Type.ValueString()}
	if !m.Agents.IsNull() && !m.Agents.IsUnknown() {
		var elems []types.Object
		diags.Append(m.Agents.ElementsAs(ctx, &elems, false)...)
		if diags.HasError() {
			return nil, diags
		}
		out.Agents = make([]apiclient.RosterEntry, 0, len(elems))
		for _, e := range elems {
			var r rosterEntryModel
			diags.Append(e.As(ctx, &r, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil, diags
			}
			entry := apiclient.RosterEntry{Type: r.Type.ValueString()}
			if !r.Id.IsNull() && !r.Id.IsUnknown() {
				v := r.Id.ValueString()
				entry.Id = &v
			}
			if !r.Version.IsNull() && !r.Version.IsUnknown() {
				v := int(r.Version.ValueInt64())
				entry.Version = &v
			}
			out.Agents = append(out.Agents, entry)
		}
	}
	return &out, diags
}
