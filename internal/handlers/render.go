package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

// Renderer parses each page template together with the shared layout and
// component partials, then caches the result by page name. Every page
// automatically gets the nav/footer chrome and access to shared components
// like {{template "job-card" .}}.
type Renderer struct {
	templatesDir string
	cache        map[string]*template.Template
}

func NewRenderer(templatesDir string) (*Renderer, error) {
	r := &Renderer{templatesDir: templatesDir, cache: map[string]*template.Template{}}
	if err := r.loadAll(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Renderer) loadAll() error {
	pages, err := filepath.Glob(filepath.Join(r.templatesDir, "pages", "*.html"))
	if err != nil {
		return err
	}
	layouts, err := filepath.Glob(filepath.Join(r.templatesDir, "layouts", "*.html"))
	if err != nil {
		return err
	}
	components, err := filepath.Glob(filepath.Join(r.templatesDir, "components", "*.html"))
	if err != nil {
		return err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		files := append([]string{page}, layouts...)
		files = append(files, components...)

		tmpl, err := template.New(name).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", name, err)
		}
		r.cache[name] = tmpl
	}
	return nil
}

// RenderBlock executes a named block within a page's template set (e.g. a
// fragment like "job-results" or "ats-columns") instead of the full "base"
// layout, for HTMX endpoints that only replace part of the page.
func (r *Renderer) RenderBlock(w http.ResponseWriter, page, block string, data any) {
	tmpl, ok := r.cache[page]
	if !ok {
		http.Error(w, fmt.Sprintf("template %s not found", page), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, block, data); err != nil {
		fmt.Fprintf(w, "<!-- render error: %v -->", err)
	}
}
func (r *Renderer) Render(w http.ResponseWriter, status int, page string, data any) {
	tmpl, ok := r.cache[page]
	if !ok {
		http.Error(w, fmt.Sprintf("template %s not found", page), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		// Headers are already sent at this point; log-and-swallow is the
		// pragmatic option rather than attempting a second write.
		fmt.Fprintf(w, "<!-- render error: %v -->", err)
	}
}
