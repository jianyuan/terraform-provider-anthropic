package provider

import (
	"context"

	"github.com/frank-bee/terraform-provider-anthropic/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DeploymentScheduleModel is the optional `schedule` block. Expression/Timezone
// are user-settable; LastRunAt/UpcomingRunsAt are server-computed.
type DeploymentScheduleModel struct {
	Expression     types.String `tfsdk:"expression"`
	Timezone       types.String `tfsdk:"timezone"`
	LastRunAt      types.String `tfsdk:"last_run_at"`
	UpcomingRunsAt types.List   `tfsdk:"upcoming_runs_at"`
}

type DeploymentModel struct {
	Id            types.String             `tfsdk:"id"`
	Name          types.String             `tfsdk:"name"`
	Description   types.String             `tfsdk:"description"`
	AgentId       types.String             `tfsdk:"agent_id"`
	AgentVersion  types.String             `tfsdk:"agent_version"`
	EnvironmentId types.String             `tfsdk:"environment_id"`
	VaultIds      types.List               `tfsdk:"vault_ids"`
	Metadata      types.Map                `tfsdk:"metadata"`
	InitialEvents types.List               `tfsdk:"initial_events"`
	Resources     types.List               `tfsdk:"resources"`
	Schedule      *DeploymentScheduleModel `tfsdk:"schedule"`
	Status        types.String             `tfsdk:"status"`
	ArchivedAt    types.String             `tfsdk:"archived_at"`
	CreatedAt     types.String             `tfsdk:"created_at"`
	UpdatedAt     types.String             `tfsdk:"updated_at"`
}

func (m *DeploymentModel) Fill(ctx context.Context, d apiclient.Deployment) error {
	m.Id = types.StringValue(d.Id)
	m.Name = types.StringValue(d.Name)
	m.EnvironmentId = types.StringValue(d.EnvironmentId)

	// description: API returns "" rather than omitting the field, treat as null
	// so an unset attribute doesn't show as a permanent diff (same convention
	// as EnvironmentModel.Fill).
	if d.Description != nil && *d.Description != "" {
		m.Description = types.StringValue(*d.Description)
	} else {
		m.Description = types.StringNull()
	}

	if d.Agent != nil {
		m.AgentId = types.StringValue(d.Agent.Id)
		m.AgentVersion = types.StringValue(d.Agent.Version)
	} else {
		m.AgentId = types.StringNull()
		m.AgentVersion = types.StringNull()
	}

	m.VaultIds = stringListOrNull(d.VaultIds)

	if d.Metadata != nil && len(*d.Metadata) > 0 {
		elems := make(map[string]attr.Value, len(*d.Metadata))
		for k, v := range *d.Metadata {
			elems[k] = types.StringValue(v)
		}
		m.Metadata = types.MapValueMust(types.StringType, elems)
	} else {
		m.Metadata = types.MapNull(types.StringType)
	}

	initialEvents, err := jsonListOrNull(d.InitialEvents)
	if err != nil {
		return err
	}
	m.InitialEvents = initialEvents

	resources, err := jsonListOrNull(d.Resources)
	if err != nil {
		return err
	}
	m.Resources = resources

	if d.Schedule != nil {
		m.Schedule = &DeploymentScheduleModel{
			Expression:     types.StringValue(d.Schedule.Expression),
			Timezone:       types.StringPointerValue(d.Schedule.Timezone),
			LastRunAt:      types.StringPointerValue(d.Schedule.LastRunAt),
			UpcomingRunsAt: stringListOrNull(d.Schedule.UpcomingRunsAt),
		}
	} else {
		m.Schedule = nil
	}

	m.Status = types.StringPointerValue(d.Status)
	m.ArchivedAt = types.StringPointerValue(d.ArchivedAt)
	m.CreatedAt = types.StringPointerValue(d.CreatedAt)
	m.UpdatedAt = types.StringPointerValue(d.UpdatedAt)

	return nil
}
