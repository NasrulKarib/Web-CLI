package http

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// TemplateHandler renders HTML templates from the web root directory.
type TemplateHandler struct {
	templates *template.Template
}

// NewTemplateHandler parses all .html templates from webRoot and returns a handler.
func NewTemplateHandler(webRoot string) (*TemplateHandler, error) {
	pattern := filepath.Join(webRoot, "*.html")
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	log.Printf("template: loaded templates from %s", pattern)
	return &TemplateHandler{templates: tmpl}, nil
}

// HandleIndex renders the index.html page.
func (t *TemplateHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	// ServeMux routes "/" as a catch-all; only render for exact "/" path.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		log.Printf("template: failed to render index: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
