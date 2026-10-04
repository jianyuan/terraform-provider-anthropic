package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwresource"
)

var _ resource.Resource = &ServiceAccountResource{}
var _ resource.ResourceWithConfigure = &ServiceAccountResource{}
var _ resource.ResourceWithImportState = &ServiceAccountResource{}

func NewServiceAccountResource() resource.Resource {
	return &ServiceAccountResource{}
}

type ServiceAccountResource struct {
	baseResource
}

func (r *ServiceAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *ServiceAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = serviceAccountSchema().GetResource(ctx)
}

func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	fwresource.Create(ctx, func(_ *ServiceAccountModel, body *anthropic.OrganizationServiceAccountNewParams) (*anthropic.ServiceAccount, error) {
		return r.authTokenClient.Organization.ServiceAccounts.New(ctx, *body)
	}, req, resp)
}

func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	fwresource.Read(ctx, func(data *ServiceAccountModel) (*anthropic.ServiceAccount, error) {
		return r.authTokenClient.Organization.ServiceAccounts.Get(ctx, data.Id.ValueString())
	}, req, resp)
}

func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	fwresource.Update(ctx, func(data *ServiceAccountModel, body *anthropic.OrganizationServiceAccountUpdateParams) (*anthropic.ServiceAccount, error) {
		return r.authTokenClient.Organization.ServiceAccounts.Update(ctx, data.Id.ValueString(), *body)
	}, req, resp)
}

func (r *ServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	fwresource.Delete(ctx, func(data *ServiceAccountModel) error {
		_, err := r.authTokenClient.Organization.ServiceAccounts.Archive(ctx, data.Id.ValueString())
		return err
	}, req, resp)
}

func (r *ServiceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
