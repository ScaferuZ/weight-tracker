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
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/partials/*.html"))

	return &PageHandler{
		tmpl: tmpl,
	}
}

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	if !middleware.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	h.tmpl.ExecuteTemplate(w, "home.html", nil)
}

func (h *PageHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	h.tmpl.ExecuteTemplate(w, "404.html", nil)
}