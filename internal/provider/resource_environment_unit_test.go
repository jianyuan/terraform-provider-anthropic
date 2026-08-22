package provider

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/jianyuan/terraform-provider-anthropic/internal/apiclient"
)

// fakeEnvironmentDeleter is a minimal stand-in for the generated apiclient
// that lets us assert which endpoints are called and feed back arbitrary
// response status codes.
type fakeEnvironmentDeleter struct {
	deleteResp  *apiclient.DeleteEnvironmentResponse
	deleteErr   error
	deleteCalls int

	archiveResp  *apiclient.ArchiveEnvironmentResponse
	archiveErr   error
	archiveCalls int
}

func (f *fakeEnvironmentDeleter) DeleteEnvironmentWithResponse(_ context.Context, _ string, _ ...apiclient.RequestEditorFn) (*apiclient.DeleteEnvironmentResponse, error) {
	f.deleteCalls++
	return f.deleteResp, f.deleteErr
}

func (f *fakeEnvironmentDeleter) ArchiveEnvironmentWithResponse(_ context.Context, _ string, _ ...apiclient.RequestEditorFn) (*apiclient.ArchiveEnvironmentResponse, error) {
	f.archiveCalls++
	return f.archiveResp, f.archiveErr
}

func deleteRespWithStatus(code int, body string) *apiclient.DeleteEnvironmentResponse {
	return &apiclient.DeleteEnvironmentResponse{
		HTTPResponse: &http.Response{StatusCode: code},
		Body:         []byte(body),
	}
}

func archiveRespWithStatus(code int, body string) *apiclient.ArchiveEnvironmentResponse {
	return &apiclient.ArchiveEnvironmentResponse{
		HTTPResponse: &http.Response{StatusCode: code},
		Body:         []byte(body),
	}
}

// Regression test for codex review P2: Delete must only fall back to Archive
// when the API rejects the hard-delete with 409 Conflict (sessions still
// reference the environment). On any other non-200/non-404 status — most
// importantly transient 5xx — we must NOT archive, since that would silently
// abandon the resource in an archived state and remove it from Terraform
// state.
func TestDeleteEnvironmentWithFallback(t *testing.T) {
	t.Run("200 OK no fallback", func(t *testing.T) {
		f := &fakeEnvironmentDeleter{deleteResp: deleteRespWithStatus(http.StatusOK, "")}
		if err := deleteEnvironmentWithFallback(t.Context(), f, "env_01"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.deleteCalls != 1 {
			t.Errorf("expected 1 delete call, got %d", f.deleteCalls)
		}
		if f.archiveCalls != 0 {
			t.Errorf("expected 0 archive calls on 200, got %d", f.archiveCalls)
		}
	})

	t.Run("404 Not Found no fallback", func(t *testing.T) {
		f := &fakeEnvironmentDeleter{deleteResp: deleteRespWithStatus(http.StatusNotFound, "")}
		if err := deleteEnvironmentWithFallback(t.Context(), f, "env_01"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.archiveCalls != 0 {
			t.Errorf("expected 0 archive calls on 404, got %d", f.archiveCalls)
		}
	})

	t.Run("409 Conflict triggers archive", func(t *testing.T) {
		f := &fakeEnvironmentDeleter{
			deleteResp:  deleteRespWithStatus(http.StatusConflict, `{"error":"sessions reference this environment"}`),
			archiveResp: archiveRespWithStatus(http.StatusOK, ""),
		}
		if err := deleteEnvironmentWithFallback(t.Context(), f, "env_01"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.deleteCalls != 1 {
			t.Errorf("expected 1 delete call, got %d", f.deleteCalls)
		}
		if f.archiveCalls != 1 {
			t.Errorf("expected 1 archive call after 409, got %d", f.archiveCalls)
		}
	})

	t.Run("500 internal error must NOT archive", func(t *testing.T) {
		f := &fakeEnvironmentDeleter{
			deleteResp: deleteRespWithStatus(http.StatusInternalServerError, "transient backend error"),
			// archiveResp set to OK to prove that even if archive *would* succeed,
			// we don't call it on 5xx.
			archiveResp: archiveRespWithStatus(http.StatusOK, ""),
		}
		err := deleteEnvironmentWithFallback(t.Context(), f, "env_01")
		if err == nil {
			t.Fatal("expected error on 500, got nil")
		}
		if f.archiveCalls != 0 {
			t.Errorf("expected 0 archive calls on 500 (regression), got %d", f.archiveCalls)
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("expected error to mention status 500, got: %v", err)
		}
	})

	t.Run("403 Forbidden must NOT archive", func(t *testing.T) {
		f := &fakeEnvironmentDeleter{
			deleteResp:  deleteRespWithStatus(http.StatusForbidden, "permission denied"),
			archiveResp: archiveRespWithStatus(http.StatusOK, ""),
		}
		err := deleteEnvironmentWithFallback(t.Context(), f, "env_01")
		if err == nil {
			t.Fatal("expected error on 403, got nil")
		}
		if f.archiveCalls != 0 {
			t.Errorf("expected 0 archive calls on 403, got %d", f.archiveCalls)
		}
	})

	t.Run("409 then archive 500 surfaces archive error", func(t *testing.T) {
		f := &fakeEnvironmentDeleter{
			deleteResp:  deleteRespWithStatus(http.StatusConflict, ""),
			archiveResp: archiveRespWithStatus(http.StatusInternalServerError, "boom"),
		}
		err := deleteEnvironmentWithFallback(t.Context(), f, "env_01")
		if err == nil {
			t.Fatal("expected error when archive fails, got nil")
		}
		if f.archiveCalls != 1 {
			t.Errorf("expected 1 archive call attempt, got %d", f.archiveCalls)
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("expected error to mention archive status 500, got: %v", err)
		}
	})
}
