package fwdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

func Read[M ModelFromAPI[T], T any](
	ctx context.Context,
	getItem func(M) (*T, error),
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data M

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := getItem(data)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewResourceReadErrorDiagnostic(err))
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *item)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
