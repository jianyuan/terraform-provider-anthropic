package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apijson"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdatasource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwtypes"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type WorkspaceRateLimitsDataSourceModel struct {
	WorkspaceID         types.String                                                                              `tfsdk:"workspace_id" apijson:",computed"`
	Query               supertypes.SingleNestedObjectValueOf[WorkspaceRateLimitsDataSourceModel_Query]            `tfsdk:"query"`
	WorkspaceRateLimits supertypes.ListNestedObjectValueOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit] `tfsdk:"workspace_rate_limits" apijson:",computed"`
}

func (m *WorkspaceRateLimitsDataSourceModel) FromAPI(ctx context.Context, data []anthropic.WorkspaceRateLimit) (diags diag.Diagnostics) {
	err := apijson.DecodeComputed(ctx, data, m, apijson.WithUnknownAsNull())
	if err != nil {
		diags.AddError("Failed to decode WorkspaceRateLimits", err.Error())
	}
	return
}

func (m *WorkspaceRateLimitsDataSourceModel) ToListParams(ctx context.Context) (*anthropic.OrganizationWorkspaceRateLimitListParams, diag.Diagnostics) {
	var diags diag.Diagnostics
	params := anthropic.OrganizationWorkspaceRateLimitListParams{}

	if fwtypes.IsKnown(m.Query) {
		query := fwdiag.Merge(m.Query.Get(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}

		if fwtypes.IsKnown(query.GroupType) {
			params.GroupType = anthropic.OrganizationWorkspaceRateLimitListParamsGroupType(query.GroupType.ValueString())
		}

		if fwtypes.IsKnown(query.IncludeInherited) {
			params.IncludeInherited = anthropic.Bool(query.IncludeInherited.ValueBool())
		}
	}

	return &params, diags
}

type WorkspaceRateLimitsDataSourceModel_Query struct {
	GroupType        types.String `tfsdk:"group_type" apijson:",computed"`
	IncludeInherited types.Bool   `tfsdk:"include_inherited" apijson:",computed"`
}

type WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit struct {
	Group       supertypes.SingleNestedObjectValueOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Group] `tfsdk:"group" apijson:",computed"`
	Limits      supertypes.ListNestedObjectValueOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit]   `tfsdk:"limits" apijson:",computed"`
	Models      supertypes.ListValueOf[string]                                                                    `tfsdk:"models" apijson:",computed"`
	RateLimitID types.String                                                                                      `tfsdk:"rate_limit_id" apijson:",computed"`
	WorkspaceID types.String                                                                                      `tfsdk:"workspace_id" apijson:",computed"`
}

func (m *WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit) FromAPI(ctx context.Context, data anthropic.WorkspaceRateLimit) (diags diag.Diagnostics) {
	m.Group = (func() supertypes.SingleNestedObjectValueOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Group] {
		var mm WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Group
		diags.Append(mm.FromAPI(ctx, data.Group)...)
		return supertypes.NewSingleNestedObjectValueOf(ctx, &mm)
	})()
	m.Limits = supertypes.NewListNestedObjectValueOfValueSlice(ctx, lo.Map(data.Limits, func(item anthropic.WorkspaceRateLimitValue, _ int) WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit {
		var mm WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit
		diags.Append(mm.FromAPI(ctx, item)...)
		return mm
	}))
	m.Models = fwtypes.ListStringValue(ctx, data.Models, data.JSON.Models)
	m.RateLimitID = fwtypes.StringValue(data.RateLimitID, data.JSON.RateLimitID)
	m.WorkspaceID = fwtypes.StringValue(data.WorkspaceID, data.JSON.WorkspaceID)
	return
}

type WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Group struct {
	Type        types.String `tfsdk:"type" apijson ",computed"`
	ID          types.String `tfsdk:"id" apijson ",computed"`
	DisplayName types.String `tfsdk:"display_name" apijson ",computed"`
}

func (m *WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Group) FromAPI(ctx context.Context, data anthropic.WorkspaceRateLimitGroupUnion) (diags diag.Diagnostics) {
	m.Type = fwtypes.StringValue(data.Type, data.JSON.Type)
	m.ID = fwtypes.StringValue(data.ID, data.JSON.ID)
	m.DisplayName = fwtypes.StringValue(data.DisplayName, data.JSON.DisplayName, fwtypes.WithUnknownAsNull())
	return
}

type WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit struct {
	Type     types.String                                                                                             `tfsdk:"type" apijson:",computed"`
	OrgLimit types.Int64                                                                                              `tfsdk:"org_limit" apijson:",computed"`
	Source   supertypes.SingleNestedObjectValueOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit_Source] `tfsdk:"source" apijson:",computed"`
	Value    types.Int64                                                                                              `tfsdk:"value" apijson:",computed"`
}

func (m *WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit) FromAPI(ctx context.Context, data anthropic.WorkspaceRateLimitValue) (diags diag.Diagnostics) {
	m.Type = fwtypes.StringValue(data.Type, data.JSON.Type)
	m.OrgLimit = fwtypes.Int64Value(data.OrgLimit, data.JSON.OrgLimit)
	m.Source = (func() supertypes.SingleNestedObjectValueOf[WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit_Source] {
		var mm WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit_Source
		diags.Append(mm.FromAPI(ctx, data.Source)...)
		return supertypes.NewSingleNestedObjectValueOf(ctx, &mm)
	})()
	m.Value = fwtypes.Int64Value(data.Value, data.JSON.Value)
	return
}

type WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit_Source struct {
	Type types.String `tfsdk:"type" apijson:",computed"`
}

func (m *WorkspaceRateLimitsDataSourceModel_WorkspaceRateLimit_Limit_Source) FromAPI(ctx context.Context, data anthropic.WorkspaceRateLimitValueSourceUnion) (diags diag.Diagnostics) {
	m.Type = fwtypes.StringValue(data.Type, data.JSON.Type)
	return
}

var _ datasource.DataSource = &WorkspaceRateLimitsDataSource{}
var _ datasource.DataSourceWithConfigure = &WorkspaceRateLimitsDataSource{}

func NewWorkspaceRateLimitsDataSource() datasource.DataSource {
	return &WorkspaceRateLimitsDataSource{}
}

type WorkspaceRateLimitsDataSource struct {
	baseDataSource
}

func (d *WorkspaceRateLimitsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_rate_limits"
}

func (d *WorkspaceRateLimitsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = workspaceRateLimitsSchema().GetDataSource(ctx)
}

func (d *WorkspaceRateLimitsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	fwdatasource.List(ctx, func(data *WorkspaceRateLimitsDataSourceModel) (fwdatasource.AutoPager[anthropic.WorkspaceRateLimit], diag.Diagnostics) {
		var diags diag.Diagnostics
		params := fwdiag.Merge(data.ToListParams(ctx))(&diags)
		if diags.HasError() {
			return nil, diags
		}
		return d.apiKeyClient.Organization.Workspaces.RateLimits.ListAutoPaging(
			ctx,
			data.WorkspaceID.ValueString(),
			*params,
		), diags
	}, req, resp)
}
