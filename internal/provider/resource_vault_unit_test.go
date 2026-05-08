package provider

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

type fakeVaultDeleter struct {
	resp  *apiclient.DeleteVaultResponse
	err   error
	calls int
}

func (f *fakeVaultDeleter) DeleteVaultWithResponse(_ context.Context, _ string, _ ...apiclient.RequestEditorFn) (*apiclient.DeleteVaultResponse, error) {
	f.calls++
	return f.resp, f.err
}

func vaultDeleteRespWithStatus(code int, body string) *apiclient.DeleteVaultResponse {
	return &apiclient.DeleteVaultResponse{
		HTTPResponse: &http.Response{StatusCode: code},
		Body:         []byte(body),
	}
}

// Regression test for codex review round 2 P1: a transient 5xx or a 4xx
// authorisation error from DELETE /v1/vaults/{id} must surface as an error
// rather than silently archiving (and dropping the vault from Terraform
// state).
func TestDeleteVault(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		f := &fakeVaultDeleter{resp: vaultDeleteRespWithStatus(http.StatusOK, "")}
		if err := deleteVault(context.Background(), f, "vlt_01"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.calls != 1 {
			t.Errorf("expected 1 delete call, got %d", f.calls)
		}
	})

	t.Run("404 Not Found is benign", func(t *testing.T) {
		f := &fakeVaultDeleter{resp: vaultDeleteRespWithStatus(http.StatusNotFound, "")}
		if err := deleteVault(context.Background(), f, "vlt_01"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("500 Internal returns error", func(t *testing.T) {
		f := &fakeVaultDeleter{resp: vaultDeleteRespWithStatus(http.StatusInternalServerError, "transient")}
		err := deleteVault(context.Background(), f, "vlt_01")
		if err == nil {
			t.Fatal("expected error on 500, got nil")
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("expected error to mention status 500, got: %v", err)
		}
	})

	t.Run("403 Forbidden returns error", func(t *testing.T) {
		f := &fakeVaultDeleter{resp: vaultDeleteRespWithStatus(http.StatusForbidden, "denied")}
		err := deleteVault(context.Background(), f, "vlt_01")
		if err == nil {
			t.Fatal("expected error on 403, got nil")
		}
		if !strings.Contains(err.Error(), "403") {
			t.Errorf("expected error to mention status 403, got: %v", err)
		}
	})
}
