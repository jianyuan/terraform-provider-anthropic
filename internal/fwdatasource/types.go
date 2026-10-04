package fwdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type AutoPager[T any] interface {
	Next() bool
	Current() T
	Err() error
}

type ModelFromAPI[T any] interface {
	FromAPI(ctx context.Context, item T) diag.Diagnostics
}

type ModelFromAPIList[T any] interface {
	FromAPI(ctx context.Context, items []T) diag.Diagnostics
}
