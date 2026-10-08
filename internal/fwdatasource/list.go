package fwdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

func List[M ModelFromAPIList[T], T any](
	ctx context.Context,
	getAutoPager func(M) (AutoPager[T], diag.Diagnostics),
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data M

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var items []T
	iter := fwdiag.Merge(getAutoPager(data))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
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
