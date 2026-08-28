package fwdiag

import "github.com/hashicorp/terraform-plugin-framework/diag"

var ErrorDiagnosticNotFound = diag.NewErrorDiagnostic("Client error", "Unable to read, got status code: 404")
