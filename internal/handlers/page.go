package handlers

import (
	"html/template"
	"net/http"
	"weight-tracker/internal/middleware"
)

type PageHandler struct {
	tmpl *template.Template
}

func NewPageHandler() *PageHandler {
	// Parse layout template first
	layoutContent, err := template.ParseFiles("templates/layout.html")
	if err != nil {
		panic(err)
	}

	// Parse home template
	homeContent, err := template.ParseFiles("templates/home.html")
	if err != nil {
		panic(err)
	}

	// Parse partials
	partials, err := template.ParseGlob("templates/partials/*.html")
	if err != nil {
		panic(err)
	}

	// Add all templates to the main template
	for _, t := range homeContent.Templates() {
		layoutContent.AddParseTree(t.Name(), t.Tree)
	}
	for _, t := range partials.Templates() {
		layoutContent.AddParseTree(t.Name(), t.Tree)
	}

	return &PageHandler{
		tmpl: layoutContent,
	}
}

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	if !middleware.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Title": "Home",
	}
	h.tmpl.ExecuteTemplate(w, "base", data)
}

func (h *PageHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	h.tmpl.ExecuteTemplate(w, "404.html", nil)
}