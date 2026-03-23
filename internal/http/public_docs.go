package httpapi

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
	"time"

	"docs.svc.plus/internal/content"
)

type publicDocsTemplateData struct {
	Title             string
	Description       string
	CanonicalURL      string
	GeneratedAt       string
	Heading           string
	Subheading        string
	IntroHTML         template.HTML
	ArticleHTML       template.HTML
	Collections       []content.DocCollection
	Navigation        content.DocsNavigation
	Breadcrumbs       []content.Crumb
	TOC               []content.TOCItem
	CurrentCollection string
	CurrentPath       string
}

var publicDocsTemplate = template.Must(template.New("public-docs").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{ .Title }}</title>
    <meta name="description" content="{{ .Description }}">
    <link rel="canonical" href="{{ .CanonicalURL }}">
    <style>
      :root {
        color-scheme: light;
        --bg: #f5f7f4;
        --panel: rgba(255,255,255,0.92);
        --border: rgba(14,24,19,0.10);
        --text: #14211a;
        --muted: #536258;
        --accent: #1f7a4f;
        --accent-soft: rgba(31,122,79,0.10);
        --shadow: 0 20px 48px rgba(9, 19, 13, 0.08);
      }
      * { box-sizing: border-box; }
      body {
        margin: 0;
        font-family: ui-serif, Georgia, Cambria, "Times New Roman", Times, serif;
        background:
          radial-gradient(circle at top left, rgba(166, 214, 190, 0.28), transparent 28rem),
          linear-gradient(180deg, #f9fbf8 0%, var(--bg) 100%);
        color: var(--text);
      }
      a { color: var(--accent); text-decoration: none; }
      a:hover { text-decoration: underline; }
      .shell { max-width: 1200px; margin: 0 auto; padding: 2rem 1.25rem 4rem; }
      .hero, .panel {
        background: var(--panel);
        border: 1px solid var(--border);
        border-radius: 1.25rem;
        box-shadow: var(--shadow);
      }
      .hero { padding: 2rem; margin-bottom: 1.5rem; }
      .eyebrow {
        margin: 0 0 0.75rem;
        font-family: ui-sans-serif, system-ui, sans-serif;
        font-size: 0.75rem;
        font-weight: 700;
        letter-spacing: 0.18em;
        text-transform: uppercase;
        color: var(--muted);
      }
      h1 {
        margin: 0;
        font-size: clamp(2rem, 5vw, 3.5rem);
        line-height: 1;
        letter-spacing: -0.05em;
      }
      .subheading {
        margin: 1rem 0 0;
        max-width: 54rem;
        font-family: ui-sans-serif, system-ui, sans-serif;
        font-size: 1rem;
        line-height: 1.7;
        color: var(--muted);
      }
      .content {
        display: grid;
        gap: 1.5rem;
      }
      .home-grid {
        display: grid;
        gap: 1rem;
        grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr));
      }
      .card {
        display: block;
        padding: 1.1rem 1.2rem;
        border: 1px solid var(--border);
        border-radius: 1rem;
        background: rgba(255,255,255,0.92);
      }
      .card:hover {
        border-color: rgba(31,122,79,0.28);
        background: #fff;
      }
      .card h2, .card h3 {
        margin: 0 0 0.45rem;
        font-size: 1.1rem;
        line-height: 1.35;
      }
      .card p, .meta, .breadcrumbs {
        margin: 0;
        font-family: ui-sans-serif, system-ui, sans-serif;
        color: var(--muted);
      }
      .layout {
        display: grid;
        gap: 1.5rem;
        grid-template-columns: minmax(0, 1fr);
      }
      .layout article,
      .layout aside {
        min-width: 0;
      }
      .panel { padding: 1.5rem; }
      .breadcrumbs {
        display: flex;
        flex-wrap: wrap;
        gap: 0.6rem;
        font-size: 0.9rem;
      }
      .breadcrumbs span::after {
        content: "›";
        margin-left: 0.6rem;
        color: rgba(83,98,88,0.5);
      }
      .breadcrumbs span:last-child::after {
        content: "";
        margin: 0;
      }
      .prose {
        font-size: 1rem;
        line-height: 1.8;
      }
      .prose h1, .prose h2, .prose h3 {
        line-height: 1.2;
        letter-spacing: -0.04em;
      }
      .prose img { max-width: 100%; border-radius: 0.9rem; }
      .prose pre {
        overflow: auto;
        padding: 1rem;
        border-radius: 1rem;
        background: #0f1f18;
        color: #edf7f0;
      }
      .prose code {
        font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
      }
      .sidebar-title {
        margin: 0 0 0.9rem;
        font-family: ui-sans-serif, system-ui, sans-serif;
        font-size: 0.78rem;
        font-weight: 700;
        letter-spacing: 0.14em;
        text-transform: uppercase;
        color: var(--muted);
      }
      .toc, .collection-list, .nav-sections {
        display: grid;
        gap: 0.6rem;
      }
      .toc a, .collection-list a, .nav-item {
        display: block;
        padding: 0.7rem 0.85rem;
        border-radius: 0.85rem;
        background: rgba(31,122,79,0.05);
        font-family: ui-sans-serif, system-ui, sans-serif;
        color: var(--text);
      }
      .toc a:hover, .collection-list a:hover, .nav-item:hover {
        background: var(--accent-soft);
        text-decoration: none;
      }
      .collection-list a.active, .nav-item.active {
        background: var(--accent-soft);
        border: 1px solid rgba(31,122,79,0.18);
      }
      .nav-section {
        display: grid;
        gap: 0.45rem;
        padding-left: 0.9rem;
        border-left: 1px solid rgba(20, 33, 26, 0.08);
      }
      .nav-section + .nav-section {
        margin-top: 1.2rem;
      }
      .nav-section-title {
        margin: 0 0 0.2rem;
        font-family: ui-sans-serif, system-ui, sans-serif;
        font-size: 0.82rem;
        font-weight: 700;
        color: var(--text);
      }
      footer {
        margin-top: 1.5rem;
        font-family: ui-sans-serif, system-ui, sans-serif;
        font-size: 0.9rem;
        color: var(--muted);
      }
      @media (min-width: 980px) {
        .layout {
          grid-template-columns: minmax(0, 1fr) 18rem;
        }
      }
    </style>
  </head>
  <body>
    <div class="shell">
      <section class="hero">
        <p class="eyebrow">docs.svc.plus</p>
        <h1>{{ .Heading }}</h1>
        <p class="subheading">{{ .Subheading }}</p>
      </section>
      {{ if .ArticleHTML }}
      <div class="layout">
        <article class="panel">
          {{ if .Breadcrumbs }}
          <nav class="breadcrumbs" aria-label="Breadcrumb">
            {{ range .Breadcrumbs }}
            <span><a href="{{ .Href }}">{{ .Label }}</a></span>
            {{ end }}
          </nav>
          {{ end }}
          <div class="prose">{{ .ArticleHTML }}</div>
        </article>
        <aside class="panel">
          {{ if .TOC }}
          <p class="sidebar-title">On this page</p>
          <nav class="toc">
            {{ range .TOC }}
            <a href="#{{ .Anchor }}">{{ .Title }}</a>
            {{ end }}
          </nav>
          {{ end }}
          {{ if .Navigation.Sections }}
          <p class="sidebar-title" style="margin-top: 1.25rem;">Browse docs</p>
          <nav class="nav-sections">
            {{ range .Navigation.Sections }}
            <section class="nav-section">
              <p class="nav-section-title">{{ .Title }}</p>
              {{ range .Items }}
              <a href="{{ .Href }}" class="nav-item{{ if eq $.CurrentPath .Href }} active{{ end }}">{{ .Title }}</a>
              {{ end }}
            </section>
            {{ end }}
          </nav>
          {{ else if .Collections }}
          <p class="sidebar-title" style="margin-top: 1.25rem;">Collections</p>
          <nav class="collection-list">
            {{ range .Collections }}
            <a href="/docs/{{ .Slug }}/{{ .DefaultVersionSlug }}" class="{{ if eq $.CurrentCollection .Slug }}active{{ end }}">{{ .Title }}</a>
            {{ end }}
          </nav>
          {{ end }}
        </aside>
      </div>
      {{ else }}
      <section class="panel">
        {{ if .IntroHTML }}
        <div class="prose">{{ .IntroHTML }}</div>
        {{ end }}
      </section>
      <section class="panel">
        {{ if .Navigation.Sections }}
        <p class="sidebar-title">Browse docs</p>
        <div class="home-grid">
          {{ range .Navigation.Sections }}
          <section class="card">
            <h2>{{ .Title }}</h2>
            <div class="nav-sections">
              <div class="nav-section" style="border-left: none; padding-left: 0;">
                {{ range .Items }}
                <a href="{{ .Href }}" class="nav-item">{{ .Title }}</a>
                {{ end }}
              </div>
            </div>
          </section>
          {{ end }}
        </div>
        {{ else }}
        <p class="sidebar-title">Browse collections</p>
        <div class="home-grid">
          {{ range .Collections }}
          <a class="card" href="/docs/{{ .Slug }}/{{ .DefaultVersionSlug }}">
            <h2>{{ .Title }}</h2>
            <p>{{ .Description }}</p>
            <p class="meta" style="margin-top: 0.85rem;">{{ len .Versions }} article(s)</p>
          </a>
          {{ end }}
        </div>
        {{ end }}
      </section>
      {{ end }}
      <footer>Generated {{ .GeneratedAt }}</footer>
    </div>
  </body>
</html>`))

func (a *App) handlePublicDocs(w http.ResponseWriter, r *http.Request) {
	trimmedPath := strings.TrimPrefix(r.URL.Path, "/docs")
	switch strings.Trim(trimmedPath, "/") {
	case "":
		a.renderPublicDocsHome(w, r)
		return
	}

	parts := strings.Split(strings.Trim(trimmedPath, "/"), "/")
	if len(parts) == 1 {
		a.redirectCollectionHome(w, r, parts[0])
		return
	}

	a.renderPublicDocPage(w, r, parts[0], strings.Join(parts[1:], "/"))
}

func (a *App) renderPublicDocsHome(w http.ResponseWriter, r *http.Request) {
	lang := resolveLang(r)
	snapshot := a.GetSnapshot()
	home, ok := snapshot.DocsHomeByLang[lang]
	if !ok {
		home = snapshot.DocsHomeByLang["default"]
	}
	navigation := pickDocsNavigation(snapshot, lang)

	a.renderPublicDocsTemplate(w, http.StatusOK, publicDocsTemplateData{
		Title:        home.Title + " | docs.svc.plus",
		Description:  home.Description,
		CanonicalURL: publicCanonicalURL(r, "/docs"),
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		Heading:      home.Title,
		Subheading:   home.Description,
		IntroHTML:    template.HTML(home.HTML),
		Collections:  filterCollectionsByLang(snapshot.Collections, lang),
		Navigation:   navigation,
		CurrentPath:  "/docs",
	})
}

func (a *App) redirectCollectionHome(w http.ResponseWriter, r *http.Request, collection string) {
	snapshot := a.GetSnapshot()
	entry, ok := snapshot.CollectionsBySlug[collection]
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/docs/"+entry.Slug+"/"+entry.DefaultVersionSlug, http.StatusPermanentRedirect)
}

func (a *App) renderPublicDocPage(w http.ResponseWriter, r *http.Request, collection, slug string) {
	lang := resolveLang(r)
	snapshot := a.GetSnapshot()
	page, ok := resolveDocPage(snapshot, lang, collection, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}

	a.renderPublicDocsTemplate(w, http.StatusOK, publicDocsTemplateData{
		Title:             page.Version.Title + " | docs.svc.plus",
		Description:       page.Version.Description,
		CanonicalURL:      publicCanonicalURL(r, "/docs/"+collection+"/"+slug),
		GeneratedAt:       time.Now().UTC().Format(time.RFC3339),
		Heading:           page.Version.Title,
		Subheading:        page.Version.Description,
		ArticleHTML:       template.HTML(page.Version.HTML),
		Collections:       filterCollectionsByLang(snapshot.Collections, lang),
		Navigation:        pickDocsNavigation(snapshot, lang),
		Breadcrumbs:       page.Breadcrumbs,
		TOC:               page.Version.TOC,
		CurrentCollection: page.Collection.Slug,
		CurrentPath:       "/docs/" + collection + "/" + slug,
	})
}

func (a *App) renderPublicDocsTemplate(
	w http.ResponseWriter,
	status int,
	data publicDocsTemplateData,
) {
	var buf bytes.Buffer
	if err := publicDocsTemplate.Execute(&buf, data); err != nil {
		http.Error(w, "failed to render docs page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func publicCanonicalURL(r *http.Request, path string) string {
	scheme := "https"
	if r.TLS == nil {
		if forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwardedProto != "" {
			scheme = forwardedProto
		}
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		host = "docs.svc.plus"
	}
	return scheme + "://" + host + path
}

func pickDocsNavigation(snapshot *content.Snapshot, lang string) content.DocsNavigation {
	if snapshot == nil {
		return content.DocsNavigation{}
	}
	if lang != "" && lang != "default" {
		if nav, ok := snapshot.DocsNavigationByLang[lang]; ok {
			return nav
		}
	}
	if nav, ok := snapshot.DocsNavigationByLang["default"]; ok {
		return nav
	}
	return content.DocsNavigation{}
}

func resolveDocPage(snapshot *content.Snapshot, lang, collection, slug string) (content.DocPage, bool) {
	if snapshot == nil {
		return content.DocPage{}, false
	}
	if lang != "" && lang != "default" {
		if page, ok := snapshot.PagesByKey[collection+":"+lang+"::"+slug]; ok {
			return page, true
		}
	}
	page, ok := snapshot.PagesByKey[collection+"::"+slug]
	return page, ok
}
