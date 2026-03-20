package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"docs.svc.plus/internal/agent"
	"docs.svc.plus/internal/config"
	"docs.svc.plus/internal/content"
	gitsync "docs.svc.plus/internal/git"
)

type App struct {
	cfg     config.Config
	indexer *content.Indexer
	agent   *agent.Handler

	mu       sync.RWMutex
	snapshot *content.Snapshot
	loadedAt time.Time
}

func NewApp(cfg config.Config) (*App, error) {
	indexer := content.NewIndexer(cfg.KnowledgeRepoPath)
	snapshot, err := indexer.Build()
	if err != nil {
		return nil, err
	}
	app := &App{
		cfg:      cfg,
		indexer:  indexer,
		snapshot: snapshot,
		loadedAt: time.Now().UTC(),
	}
	app.agent = agent.NewHandler(app)
	return app, nil
}

func (a *App) RepoPath() string {
	return a.cfg.KnowledgeRepoPath
}

func (a *App) GetSnapshot() *content.Snapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.snapshot
}

func (a *App) Reload(pull bool) content.ReloadResult {
	result := content.ReloadResult{
		Pulled:   false,
		Reloaded: false,
		LoadedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if pull {
		ok, message, err := gitsync.Pull(a.cfg.KnowledgeRepoPath)
		result.Pulled = ok
		result.Message = message
		if err != nil {
			result.Message = err.Error()
			return result
		}
	}
	snapshot, err := a.indexer.Build()
	if err != nil {
		result.Message = err.Error()
		return result
	}
	a.mu.Lock()
	a.snapshot = snapshot
	a.loadedAt = time.Now().UTC()
	a.mu.Unlock()
	result.Reloaded = true
	return result
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/docs", a.handlePublicDocs)
	mux.HandleFunc("/docs/", a.handlePublicDocs)

	api := http.NewServeMux()
	api.HandleFunc("/api/v1/docs/home", a.handleDocsHome)
	api.HandleFunc("/api/v1/docs/collections", a.handleDocCollections)
	api.HandleFunc("/api/v1/docs/pages/", a.handleDocPage)
	api.HandleFunc("/api/v1/blogs/", a.handleBlogPost)
	api.HandleFunc("/api/v1/blogs", a.handleBlogs)
	api.HandleFunc("/api/v1/home/latest-blogs", a.handleLatestBlogs)
	api.HandleFunc("/api/v1/admin/reload", a.handleReload)
	api.HandleFunc("/api/v1/agent/invoke", a.handleAgentInvoke)

	mux.Handle("/api/", RequireServiceToken(a.cfg.InternalServiceToken, api))
	return mux
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"loadedAt": a.loadedAt.Format(time.RFC3339),
	})
}

func (a *App) handleDocsHome(w http.ResponseWriter, r *http.Request) {
	lang := resolveLang(r)
	snapshot := a.GetSnapshot()
	home, ok := snapshot.DocsHomeByLang[lang]
	if !ok {
		home = snapshot.DocsHomeByLang["default"]
	}
	writeJSON(w, http.StatusOK, home)
}

func (a *App) handleDocCollections(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.GetSnapshot().Collections)
}

func (a *App) handleDocPage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/docs/pages/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "collection_and_slug_required"})
		return
	}
	page, ok := a.GetSnapshot().PagesByKey[parts[0]+"::"+parts[1]]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (a *App) handleBlogs(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	page := parseInt(r.URL.Query().Get("page"), 1)
	pageSize := parseInt(r.URL.Query().Get("pageSize"), 10)
	if pageSize > 50 {
		pageSize = 50
	}

	filtered := make([]content.BlogPost, 0)
	for _, post := range a.GetSnapshot().Blogs {
		if category != "" && (post.Category == nil || post.Category.Key != category) {
			continue
		}
		if query != "" {
			haystack := strings.ToLower(post.Title + " " + post.Excerpt + " " + post.Plaintext)
			if !strings.Contains(haystack, strings.ToLower(query)) {
				continue
			}
		}
		filtered = append(filtered, post)
	}

	total := len(filtered)
	totalPages := 1
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	writeJSON(w, http.StatusOK, content.BlogList{
		Posts:      filtered[start:end],
		Categories: a.GetSnapshot().BlogCategories,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (a *App) handleBlogPost(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/blogs/")
	post, ok := a.GetSnapshot().BlogsBySlug[slug]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
		return
	}
	writeJSON(w, http.StatusOK, post)
}

func (a *App) handleLatestBlogs(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 7)
	if limit > 20 {
		limit = 20
	}
	posts := a.GetSnapshot().Blogs
	if len(posts) > limit {
		posts = posts[:limit]
	}
	writeJSON(w, http.StatusOK, posts)
}

func (a *App) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		return
	}
	result := a.Reload(r.URL.Query().Get("pull") != "false")
	status := http.StatusOK
	if !result.Reloaded {
		status = http.StatusBadGateway
	}
	writeJSON(w, status, result)
}

func (a *App) handleAgentInvoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		return
	}
	var req agent.InvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	result, err := a.agent.Invoke(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tool": req.Tool, "result": result})
}

func resolveLang(r *http.Request) string {
	lang := strings.TrimSpace(r.URL.Query().Get("lang"))
	if lang == "zh" || lang == "en" {
		return lang
	}
	return "default"
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
