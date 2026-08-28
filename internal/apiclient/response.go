package apiclient

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-anthropic/internal/fwdiag"
)

type JSON200Response[T any] interface {
	StatusCode() int
	GetBody() []byte
	GetJSON200() *T
}

func CreateJSON200[T any, R JSON200Response[T]](httpResp R, err error) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics
	if err != nil {
		diags.AddError("Client error", fmt.Sprintf("Unable to create, got error: %s", err))
		return nil, diags
	} else if httpResp.StatusCode() != http.StatusOK {
		diags.AddError("Client error", fmt.Sprintf("Unable to create, got status code: %d, body: %s", httpResp.StatusCode(), string(httpResp.GetBody())))
		return nil, diags
	} else if httpResp.GetJSON200() == nil {
		diags.AddError("Client error", "Unable to create, got empty response body")
		return nil, diags
	}
	return httpResp.GetJSON200(), diags
}

func ReadJSON200[T any, R JSON200Response[T]](httpResp R, err error) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics
	if err != nil {
		diags.AddError("Client error", fmt.Sprintf("Unable to read, got error: %s", err))
		return nil, diags
	} else if httpResp.StatusCode() == http.StatusNotFound {
		diags.Append(fwdiag.ErrorDiagnosticNotFound)
		return nil, diags
	} else if httpResp.StatusCode() != http.StatusOK {
		diags.AddError("Client error", fmt.Sprintf("Unable to read, got status code: %d, body: %s", httpResp.StatusCode(), string(httpResp.GetBody())))
		return nil, diags
	} else if httpResp.GetJSON200() == nil {
		diags.AddError("Client error", "Unable to read, got empty response body")
		return nil, diags
	}
	return httpResp.GetJSON200(), diags
}

func UpdateJSON200[T any, R JSON200Response[T]](httpResp R, err error) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics
	if err != nil {
		diags.AddError("Client error", fmt.Sprintf("Unable to update, got error: %s", err))
		return nil, diags
	} else if httpResp.StatusCode() != http.StatusOK {
		diags.AddError("Client error", fmt.Sprintf("Unable to update, got status code: %d, body: %s", httpResp.StatusCode(), string(httpResp.GetBody())))
		return nil, diags
	} else if httpResp.GetJSON200() == nil {
		diags.AddError("Client error", "Unable to update, got empty response body")
		return nil, diags
	}
	return httpResp.GetJSON200(), diags
}

func DeleteJSON200[T any, R JSON200Response[T]](httpResp R, err error) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics
	if err != nil {
		diags.AddError("Client error", fmt.Sprintf("Unable to delete, got error: %s", err))
		return nil, diags
	} else if httpResp.StatusCode() == http.StatusNotFound {
		// Already deleted
		return nil, diags
	} else if httpResp.StatusCode() != http.StatusOK {
		diags.AddError("Client error", fmt.Sprintf("Unable to delete, got status code: %d, body: %s", httpResp.StatusCode(), string(httpResp.GetBody())))
		return nil, diags
	} else if httpResp.GetJSON200() == nil {
		diags.AddError("Client error", "Unable to delete, got empty response body")
		return nil, diags
	}
	return httpResp.GetJSON200(), diags
}
