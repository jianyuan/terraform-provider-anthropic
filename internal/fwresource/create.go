package fwresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

func Create[M ModelForCreate[T, B], T, B any](
	ctx context.Context,
	createFunc func(M, B) (*T, error),
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data M

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := fwdiag.Merge(data.ToAPIForCreate(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := createFunc(data, body)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewResourceCreateErrorDiagnostic(err))
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *item)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
