package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"jobhoo/internal/models"
)

type Renderer struct {
	templates map[string]*template.Template
	baseDir   string
}

var funcMap = template.FuncMap{
	"employmentLabel": models.EmploymentTypeLabel,
	"arrangementLabel": models.WorkArrangementLabel,
	"stageLabel": models.StageLabel,
	"finalStatusLabel": func(s *models.ApplicationFinalStatus) string {
		if s == nil {
			return ""
		}
		return models.FinalStatusLabel(*s)
	},
	"formatDate": func(t interface{ Format(string) string }) string {
		return t.Format("2 Jan 2006")
	},
	"add": func(a, b int) int { return a + b },
	// derefFinalStatus safely unwraps a *models.ApplicationFinalStatus into a
	// plain string ("" when nil), since html/template's eq cannot compare a
	// pointer against an untyped string constant.
	"derefFinalStatus": func(s *models.ApplicationFinalStatus) string {
		if s == nil {
			return ""
		}
		return string(*s)
	},
}

// NewRenderer parses every page template together with the shared layout and
// partials, so {{define "base"}} / {{template "content"}} composition works.
func NewRenderer(baseDir string) (*Renderer, error) {
	r := &Renderer{templates: map[string]*template.Template{}, baseDir: baseDir}

	layout := filepath.Join(baseDir, "layout", "base.html")
	partials, err := filepath.Glob(filepath.Join(baseDir, "partials", "*.html"))
	if err != nil {
		return nil, err
	}
	pages, err := filepath.Glob(filepath.Join(baseDir, "pages", "*.html"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		files := append([]string{layout, page}, partials...)
		tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			return nil, err
		}
		r.templates[name] = tmpl
	}

	return r, nil
}

// Render executes the named page template within the base layout.
func (r *Renderer) Render(w http.ResponseWriter, page string, data interface{}) {
	tmpl, ok := r.templates[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("render error (%s): %v", page, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// RenderPartial executes just a named partial/fragment template, used for
// htmx responses that swap a small piece of the page.
func (r *Renderer) RenderPartial(w http.ResponseWriter, page, blockName string, data interface{}) {
	tmpl, ok := r.templates[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, blockName, data); err != nil {
		log.Printf("render partial error (%s/%s): %v", page, blockName, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
