package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"docs.svc.plus/internal/config"
)

func TestPublicDocsRoutes(t *testing.T) {
	repoPath := t.TempDir()
	mustWriteFile(t, filepath.Join(repoPath, "docs", "index.md"), "# Docs Home\n\nHello docs.")
	mustWriteFile(t, filepath.Join(repoPath, "docs", "guide", "overview.md"), "# Overview\n\nGuide body.")
	if err := os.MkdirAll(filepath.Join(repoPath, "content"), 0o755); err != nil {
		t.Fatalf("mkdir content: %v", err)
	}

	app, err := NewApp(config.Config{
		KnowledgeRepoPath: repoPath,
		Port:              "8084",
	})
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	router := app.Routes()

	t.Run("renders docs home", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Docs Home") {
			t.Fatalf("expected docs home HTML, got %q", rec.Body.String())
		}
	})

	t.Run("redirects collection to default page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/guide", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusPermanentRedirect {
			t.Fatalf("expected 308, got %d", rec.Code)
		}
		if location := rec.Header().Get("Location"); location != "/docs/guide/overview" {
			t.Fatalf("unexpected redirect location %q", location)
		}
	})

	t.Run("renders document page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/guide/overview", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Guide body.") {
			t.Fatalf("expected guide body HTML, got %q", rec.Body.String())
		}
	})
}

func mustWriteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
