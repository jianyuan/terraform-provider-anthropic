package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type VaultModel struct {
	Id          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Metadata    types.Map    `tfsdk:"metadata"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	ArchivedAt  types.String `tfsdk:"archived_at"`
}

func (m *VaultModel) Fill(ctx context.Context, v apiclient.Vault) diag.Diagnostics {
	m.Id = types.StringValue(v.Id)
	m.DisplayName = types.StringValue(v.DisplayName)
	m.CreatedAt = types.StringValue(v.CreatedAt)
	m.UpdatedAt = types.StringValue(v.UpdatedAt)
	m.ArchivedAt = types.StringPointerValue(v.ArchivedAt)

	// Empty and absent metadata are semantically the same on the API side
	// (the docs treat omitted and {} identically), so collapse both to null
	// in state. The matching schema validator rejects an explicit
	// `metadata = {}` so the user only ever has one representation for "no
	// metadata".
	if v.Metadata == nil || len(*v.Metadata) == 0 {
		m.Metadata = types.MapNull(types.StringType)
		return nil
	}

	metadata, diags := types.MapValueFrom(ctx, types.StringType, *v.Metadata)
	if diags.HasError() {
		return diags
	}
	m.Metadata = metadata
	return nil
}
