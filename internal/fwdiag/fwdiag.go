package fwdiag

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	intfwdiag "github.com/jianyuan/terraform-plugin-framework-utils/fwdiag"
)

func Merge[T any](v T, sourceDiags diag.Diagnostics) func(targetDiags *diag.Diagnostics) T {
	return intfwdiag.Merge(v, sourceDiags)
}
