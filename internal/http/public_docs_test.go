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
	mustWriteFile(t, filepath.Join(repoPath, "docs", "en", "README.md"), "---\ntitle: English Docs\ndescription: English home\n---\n# English Docs")
	mustWriteFile(t, filepath.Join(repoPath, "docs", "zh", "README.md"), "---\ntitle: 中文文档\ndescription: 中文首页\n---\n# 中文文档")
	mustWriteFile(t, filepath.Join(repoPath, "docs", "navigation.en.yaml"), "title: English Docs\ndescription: English nav\nsections:\n  - title: Get started\n    items:\n      - title: Welcome\n        href: /docs/guide/overview\n")
	mustWriteFile(t, filepath.Join(repoPath, "docs", "navigation.zh.yaml"), "title: 中文文档\ndescription: 中文导航\nsections:\n  - title: 开始使用\n    items:\n      - title: 欢迎页\n        href: /docs/guide/overview\n")
	mustWriteFile(t, filepath.Join(repoPath, "docs", "guide", "overview.md"), "# Overview\n\nGuide body.")
	mustWriteFile(t, filepath.Join(repoPath, "docs", "guide", "overview.zh.md"), "---\nslug: overview\nlang: zh\ntitle: 欢迎页\ndescription: 中文页面\n---\n# 欢迎页\n\n中文内容。")
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

	t.Run("renders localized home and page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs?lang=zh", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "中文文档") || !strings.Contains(body, "开始使用") {
			t.Fatalf("expected localized home and navigation, got %q", body)
		}

		req = httptest.NewRequest(http.MethodGet, "/docs/guide/overview?lang=zh", nil)
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body = rec.Body.String()
		if !strings.Contains(body, "中文内容。") || !strings.Contains(body, "欢迎页") {
			t.Fatalf("expected localized page, got %q", body)
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
