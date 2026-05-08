package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type EnvironmentModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Config     types.Object `tfsdk:"config"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
	ArchivedAt types.String `tfsdk:"archived_at"`
}

type environmentNetworkingModel struct {
	Type                 types.String `tfsdk:"type"`
	AllowedHosts         types.List   `tfsdk:"allowed_hosts"`
	AllowMcpServers      types.Bool   `tfsdk:"allow_mcp_servers"`
	AllowPackageManagers types.Bool   `tfsdk:"allow_package_managers"`
}

type environmentPackagesModel struct {
	Apt   types.List `tfsdk:"apt"`
	Cargo types.List `tfsdk:"cargo"`
	Gem   types.List `tfsdk:"gem"`
	Go    types.List `tfsdk:"go"`
	Npm   types.List `tfsdk:"npm"`
	Pip   types.List `tfsdk:"pip"`
}

type environmentConfigModel struct {
	Type       types.String `tfsdk:"type"`
	Networking types.Object `tfsdk:"networking"`
	Packages   types.Object `tfsdk:"packages"`
}

func environmentNetworkingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":                   types.StringType,
		"allowed_hosts":          types.ListType{ElemType: types.StringType},
		"allow_mcp_servers":      types.BoolType,
		"allow_package_managers": types.BoolType,
	}
}

func environmentPackagesAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"apt":   types.ListType{ElemType: types.StringType},
		"cargo": types.ListType{ElemType: types.StringType},
		"gem":   types.ListType{ElemType: types.StringType},
		"go":    types.ListType{ElemType: types.StringType},
		"npm":   types.ListType{ElemType: types.StringType},
		"pip":   types.ListType{ElemType: types.StringType},
	}
}

func environmentConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":       types.StringType,
		"networking": types.ObjectType{AttrTypes: environmentNetworkingAttrTypes()},
		"packages":   types.ObjectType{AttrTypes: environmentPackagesAttrTypes()},
	}
}

func stringSliceToList(ctx context.Context, in *[]string) (types.List, diag.Diagnostics) {
	if in == nil {
		return types.ListNull(types.StringType), nil
	}
	return types.ListValueFrom(ctx, types.StringType, *in)
}

// boolFromPointer turns a nullable API bool into a Terraform Bool, treating
// nil as false rather than null. Use this for fields whose API default is
// documented as false; otherwise the API omitting the field on echo would
// flip state to null and produce a perpetual diff.
func boolFromPointer(p *bool) types.Bool {
	if p == nil {
		return types.BoolValue(false)
	}
	return types.BoolValue(*p)
}

// packagesAllNil reports whether the API's Packages object exists but has
// no per-manager lists set. We treat that case as "no packages configured"
// in state so a user who omits the `packages` block doesn't see a phantom
// `{ apt = null, cargo = null, ... }` object after refresh — which would
// force a replacement plan because the enclosing `config` is RequiresReplace.
func packagesAllNil(p *apiclient.Packages) bool {
	return p.Apt == nil && p.Cargo == nil && p.Gem == nil &&
		p.Go == nil && p.Npm == nil && p.Pip == nil
}

func listToStringSlice(ctx context.Context, l types.List) (*[]string, diag.Diagnostics) {
	if l.IsNull() || l.IsUnknown() {
		return nil, nil
	}
	out := []string{}
	diags := l.ElementsAs(ctx, &out, false)
	return &out, diags
}

func (m *EnvironmentModel) Fill(ctx context.Context, e apiclient.Environment) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Id = types.StringValue(e.Id)
	m.Name = types.StringValue(e.Name)
	m.CreatedAt = types.StringValue(e.CreatedAt)
	m.UpdatedAt = types.StringValue(e.UpdatedAt)
	m.ArchivedAt = types.StringPointerValue(e.ArchivedAt)

	netModel := environmentNetworkingModel{
		Type: types.StringValue(e.Config.Networking.Type),
	}
	// allowed_hosts is Optional-only by schema design (so a user removing
	// the line actually tightens security on update); preserve null when
	// the API doesn't echo a list. Drift here would only happen if the API
	// normalizes an explicit `[]` to nil, which is the Anthropic SDK's
	// expected behaviour but not contractually guaranteed.
	if l, d := stringSliceToList(ctx, e.Config.Networking.AllowedHosts); d.HasError() {
		diags.Append(d...)
	} else {
		netModel.AllowedHosts = l
	}
	// allow_mcp_servers and allow_package_managers default to false per the
	// API docs. The API may omit them from the response (the schema marks
	// them optional), so coerce nil to false rather than letting them appear
	// as null in state — otherwise a configuration that explicitly sets
	// either to false would never round-trip and config (which is
	// RequiresReplace) would replan a replacement on every refresh.
	netModel.AllowMcpServers = boolFromPointer(e.Config.Networking.AllowMcpServers)
	netModel.AllowPackageManagers = boolFromPointer(e.Config.Networking.AllowPackageManagers)

	netObj, d := types.ObjectValueFrom(ctx, environmentNetworkingAttrTypes(), netModel)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	pkgsObj := types.ObjectNull(environmentPackagesAttrTypes())
	if e.Config.Packages != nil && !packagesAllNil(e.Config.Packages) {
		pkgs := environmentPackagesModel{}
		fields := []struct {
			dst  *types.List
			from *[]string
		}{
			{&pkgs.Apt, e.Config.Packages.Apt},
			{&pkgs.Cargo, e.Config.Packages.Cargo},
			{&pkgs.Gem, e.Config.Packages.Gem},
			{&pkgs.Go, e.Config.Packages.Go},
			{&pkgs.Npm, e.Config.Packages.Npm},
			{&pkgs.Pip, e.Config.Packages.Pip},
		}
		for _, f := range fields {
			// Preserve nil ⇄ null for symmetry with the Optional-only
			// schema; coercing absent fields to empty would make per-manager
			// lists sticky on update.
			l, d := stringSliceToList(ctx, f.from)
			diags.Append(d...)
			*f.dst = l
		}
		if diags.HasError() {
			return diags
		}
		obj, d := types.ObjectValueFrom(ctx, environmentPackagesAttrTypes(), pkgs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		pkgsObj = obj
	}

	cfg := environmentConfigModel{
		Type:       types.StringValue(e.Config.Type),
		Networking: netObj,
		Packages:   pkgsObj,
	}

	cfgObj, d := types.ObjectValueFrom(ctx, environmentConfigAttrTypes(), cfg)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Config = cfgObj

	return diags
}

// ToAPI converts the Terraform model into an API EnvironmentConfig.
func environmentConfigToAPI(ctx context.Context, cfg types.Object) (apiclient.EnvironmentConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := apiclient.EnvironmentConfig{Type: "cloud"}

	if cfg.IsNull() || cfg.IsUnknown() {
		diags.AddError("Missing config", "config is required")
		return out, diags
	}

	var c environmentConfigModel
	diags.Append(cfg.As(ctx, &c, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return out, diags
	}

	if !c.Type.IsNull() && !c.Type.IsUnknown() && c.Type.ValueString() != "" {
		out.Type = c.Type.ValueString()
	}

	var n environmentNetworkingModel
	diags.Append(c.Networking.As(ctx, &n, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return out, diags
	}

	out.Networking = apiclient.Networking{Type: n.Type.ValueString()}
	if hosts, d := listToStringSlice(ctx, n.AllowedHosts); d.HasError() {
		diags.Append(d...)
		return out, diags
	} else {
		out.Networking.AllowedHosts = hosts
	}
	if !n.AllowMcpServers.IsNull() && !n.AllowMcpServers.IsUnknown() {
		v := n.AllowMcpServers.ValueBool()
		out.Networking.AllowMcpServers = &v
	}
	if !n.AllowPackageManagers.IsNull() && !n.AllowPackageManagers.IsUnknown() {
		v := n.AllowPackageManagers.ValueBool()
		out.Networking.AllowPackageManagers = &v
	}

	if !c.Packages.IsNull() && !c.Packages.IsUnknown() {
		var p environmentPackagesModel
		diags.Append(c.Packages.As(ctx, &p, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return out, diags
		}
		pkg := apiclient.Packages{}
		fields := []struct {
			from types.List
			dst  **[]string
		}{
			{p.Apt, &pkg.Apt},
			{p.Cargo, &pkg.Cargo},
			{p.Gem, &pkg.Gem},
			{p.Go, &pkg.Go},
			{p.Npm, &pkg.Npm},
			{p.Pip, &pkg.Pip},
		}
		for _, f := range fields {
			s, d := listToStringSlice(ctx, f.from)
			if d.HasError() {
				diags.Append(d...)
				return out, diags
			}
			*f.dst = s
		}
		out.Packages = &pkg
	}

	return out, diags
}
