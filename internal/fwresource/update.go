package fwresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

func Update[M ModelForUpdate[T, B], T, B any](
	ctx context.Context,
	updateFunc func(M, B) (*T, error),
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data M

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := fwdiag.Merge(data.ToAPIForUpdate(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := updateFunc(data, body)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewResourceUpdateErrorDiagnostic(err))
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *item)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
