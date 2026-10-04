package fwresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ModelFromAPI[T any] interface {
	FromAPI(ctx context.Context, item T) diag.Diagnostics
}

type ModelForCreate[T any, B any] interface {
	ModelFromAPI[T]
	ToAPIForCreate(ctx context.Context) (B, diag.Diagnostics)
}

type ModelForUpdate[T any, B any] interface {
	ModelFromAPI[T]
	ToAPIForUpdate(ctx context.Context) (B, diag.Diagnostics)
}
