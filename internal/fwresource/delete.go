package fwresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

func Delete[M ModelFromAPI[T], T any](
	ctx context.Context,
	deleteFunc func(M) error,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data M

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := deleteFunc(data)
	if err != nil {
		if !apiclient.IsNotFoundError(err) {
			resp.Diagnostics.Append(fwdiag.NewResourceDeleteErrorDiagnostic(err))
		}
	}
}
