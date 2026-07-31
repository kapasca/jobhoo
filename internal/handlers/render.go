package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

// descriptionPolicy allows basic formatting tags used in job/company descriptions.
var descriptionPolicy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "strong", "em", "b", "i", "u", "ul", "ol", "li", "h3", "h4", "h5", "blockquote", "pre", "code")
	p.AllowAttrs("href").OnElements("a")
	p.AllowStandardURLs()
	p.RequireNoFollowOnLinks(true)
	return p
}()

// currencySymbols maps ISO 4217 codes to display symbols.
var currencySymbols = map[string]string{
	"IDR": "Rp", "SGD": "S$", "MYR": "RM", "THB": "฿",
	"VND": "₫", "PHP": "₱", "USD": "$", "AUD": "A$", "NZD": "NZ$",
}

// thousandsSep formats n as an integer string with comma separators.
func thousandsSep(n int) string {
	s := strconv.Itoa(n)
	out := make([]byte, 0, len(s)+(len(s)-1)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

var templateFuncs = template.FuncMap{
	// truncate cuts s to at most n runes and appends "..." if it was longer.
	"truncate": func(n int, s string) string {
		if utf8.RuneCountInString(s) <= n {
			return s
		}
		runes := []rune(s)
		return string(runes[:n]) + "..."
	},
	// safeHTML sanitizes s and returns trusted HTML safe to render unescaped.
	"safeHTML": func(s string) template.HTML {
		return template.HTML(descriptionPolicy.Sanitize(s))
	},
	// fmtDate formats a nullable time pointer as "2 Jan 2006" or "—" if nil.
	"fmtDate": func(t *time.Time) string {
		if t == nil {
			return "—"
		}
		return t.Format("2 Jan 2006")
	},
	// fmtSalary formats a nullable salary integer with currency symbol and
	// thousands separators (e.g. *5000000, "IDR" → "Rp 5,000,000").
	"fmtSalary": func(amount *int, currency string) string {
		if amount == nil {
			return ""
		}
		sym, ok := currencySymbols[currency]
		if !ok {
			sym = currency
		}
		return fmt.Sprintf("%s %s", sym, thousandsSep(*amount))
	},
}

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

		tmpl, err := template.New(name).Funcs(templateFuncs).ParseFiles(files...)
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
