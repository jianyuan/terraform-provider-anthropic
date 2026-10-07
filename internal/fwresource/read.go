package fwresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

func Read[M ModelFromAPI[T], T any](
	ctx context.Context,
	getItem func(M) (*T, error),
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data M

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := getItem(data)
	if err != nil {
		if apiclient.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.Append(fwdiag.NewResourceReadErrorDiagnostic(err))
		}
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *item)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
