package fwdiag

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	intfwdiag "github.com/jianyuan/terraform-plugin-framework-utils/fwdiag"
)

func Merge[T any](v T, sourceDiags diag.Diagnostics) func(targetDiags *diag.Diagnostics) T {
	return intfwdiag.Merge(v, sourceDiags)
}

func NewResourceCreateErrorDiagnostic(err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic("Creating Resource", err.Error())
}

func NewResourceUpdateErrorDiagnostic(err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic("Updating Resource", err.Error())
}

func NewResourceDeleteErrorDiagnostic(err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic("Deleting Resource", err.Error())
}

func NewResourceReadErrorDiagnostic(err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic("Reading Resource", err.Error())
}

func NewResourceListErrorDiagnostic(err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic("Listing Resources", err.Error())
}
