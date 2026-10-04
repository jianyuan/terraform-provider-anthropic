package fwdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

func List[M ModelFromAPIList[T], T any](
	ctx context.Context,
	getAutoPager func(M) AutoPager[T],
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data M

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var items []T
	iter := getAutoPager(data)
	for iter.Next() {
		items = append(items, iter.Current())
	}
	if err := iter.Err(); err != nil {
		resp.Diagnostics.Append(fwdiag.NewResourceListErrorDiagnostic(err))
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, items)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
