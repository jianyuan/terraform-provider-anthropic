package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

// metadataValueValidators is shared between vault and agent metadata
// attributes. Two checks combined:
//
//   - SizeAtLeast(1): the API treats `metadata = {}` and an absent metadata
//     field identically, so allowing both representations in config produces
//     non-convergent state. Forbid the empty-map form; users who want no
//     metadata simply omit the attribute.
//   - LengthAtLeast(1) on values: the API uses an empty-string value as the
//     tombstone in update merges (set a key to "" to delete it), so a literal
//     "" value would be indistinguishable from a deletion.
func metadataValueValidators() []validator.Map {
	return []validator.Map{
		mapvalidator.SizeAtLeast(1),
		mapvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
	}
}

func NewVaultResource() resource.Resource {
	return &VaultResource{}
}

var _ resource.Resource = &VaultResource{}
var _ resource.ResourceWithImportState = &VaultResource{}

type VaultResource struct {
	baseResource
}

func (r *VaultResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault"
}

func (r *VaultResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Vault resource. A vault is a per-end-user collection of credentials for MCP authentication used by Managed Agents sessions. " +
			"Credentials themselves are not managed by this resource and must be created out-of-band.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Vault.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Human-readable name for the Vault.",
				Required:            true,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Arbitrary key-value pairs for tracking. Removed keys are sent to the API as empty strings to delete them; you cannot store a literal empty string as a value.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators:          metadataValueValidators(),
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Vault was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Vault was last updated.",
				Computed:            true,
			},
			"archived_at": schema.StringAttribute{
				MarkdownDescription: "RFC 3339 datetime string indicating when the Vault was archived, or null if not archived.",
				Computed:            true,
			},
		},
	}
}

func vaultMetadataFromMap(ctx context.Context, m types.Map) (*map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.IsNull() || m.IsUnknown() {
		return nil, diags
	}
	out := map[string]string{}
	diags.Append(m.ElementsAs(ctx, &out, false)...)
	return &out, diags
}

// metadataDiff returns a metadata payload suitable for an update call: every
// new/changed key carries its new value, and every key removed in the plan is
// sent with an empty string (the API's tombstone for "delete this key"). When
// the plan and state both decode to empty maps but are not Equal at the
// Terraform level (e.g. plan = {}, state = null), an empty payload is
// returned so the caller can still send `metadata: {}` explicitly to give
// convergence a chance.
func metadataDiff(ctx context.Context, state, plan types.Map) (*map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	stateMap := map[string]string{}
	if !state.IsNull() && !state.IsUnknown() {
		diags.Append(state.ElementsAs(ctx, &stateMap, false)...)
	}

	planMap := map[string]string{}
	if !plan.IsNull() && !plan.IsUnknown() {
		diags.Append(plan.ElementsAs(ctx, &planMap, false)...)
	}

	if diags.HasError() {
		return nil, diags
	}

	out := map[string]string{}
	for k, v := range planMap {
		out[k] = v
	}
	for k := range stateMap {
		if _, kept := planMap[k]; !kept {
			out[k] = ""
		}
	}
	return &out, diags
}

func (r *VaultResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VaultModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.CreateVaultJSONRequestBody{
		DisplayName: data.DisplayName.ValueString(),
	}

	metadata, mDiags := vaultMetadataFromMap(ctx, data.Metadata)
	resp.Diagnostics.Append(mDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	body.Metadata = metadata

	httpResp, err := r.client.CreateVaultWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create vault, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create vault, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create vault, got empty response body")
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VaultResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VaultModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetVaultWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read vault, got error: %s", err))
		return
	}
	if httpResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read vault, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to read vault, got empty response body")
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VaultResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state VaultModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.UpdateVaultJSONRequestBody{}

	if !plan.DisplayName.Equal(state.DisplayName) {
		v := plan.DisplayName.ValueString()
		body.DisplayName = &v
	}

	if !plan.Metadata.Equal(state.Metadata) {
		metadata, mDiags := metadataDiff(ctx, state.Metadata, plan.Metadata)
		resp.Diagnostics.Append(mDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		body.Metadata = metadata
	}

	httpResp, err := r.client.UpdateVaultWithResponse(ctx, state.Id.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update vault, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update vault, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
	if httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update vault, got empty response body")
		return
	}

	resp.Diagnostics.Append(plan.Fill(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VaultResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VaultModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := deleteVault(ctx, r.client, data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
	}
}

// deleteVault hard-deletes the vault. Per the API spec vaults are hard-delete
// only — there is no documented 409/conflict fallback like environments have,
// so we surface any non-200/non-404 status as an error rather than silently
// archiving (which would lose the vault from Terraform state while leaving it
// in a half-deleted form on the API side).
func deleteVault(ctx context.Context, client vaultDeleter, id string) error {
	httpResp, err := client.DeleteVaultWithResponse(ctx, id)
	if err != nil {
		return fmt.Errorf("Unable to delete vault, got error: %s", err)
	}
	switch httpResp.StatusCode() {
	case http.StatusOK, http.StatusNotFound:
		return nil
	}
	return fmt.Errorf("Unable to delete vault, got status code %d: %s",
		httpResp.StatusCode(), string(httpResp.Body))
}

// vaultDeleter is the subset of *apiclient.ClientWithResponses that
// deleteVault uses. Splitting it out lets the unit test substitute a fake
// without spinning up a full provider.
type vaultDeleter interface {
	DeleteVaultWithResponse(ctx context.Context, vaultId string, reqEditors ...apiclient.RequestEditorFn) (*apiclient.DeleteVaultResponse, error)
}

func (r *VaultResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
