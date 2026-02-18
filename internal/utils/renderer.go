package utils

import (
	"net/http"
	"text/template"
)

var templates map[string]*template.Template

// InitTemplates for a one-time call in the main file and caching of templates.
func InitTemplates() {
	templates = make(map[string]*template.Template)
	pages := []string{"home.html", "keys.html", "login.html", "profile.html", "4xx.html", "5xx.html"}

	for _, page := range pages {
		// Parse once at startup
		t, err := template.ParseFiles("templates/base.html", "templates/"+page)
		if err != nil {
			panic("Error loading template " + page + ": " + err.Error())
		}
		templates[page] = t
	}
}

// RenderTemplate renders the template using the base.
func RenderTemplate(w http.ResponseWriter, page string, data map[string]interface{}, r *http.Request) {
	tmpl, ok := templates[page]
	if !ok {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	csrfToken, _ := r.Context().Value("csrf_token").(string)
	if data == nil {
		data = make(map[string]interface{})
	}
	data["CSRFToken"] = csrfToken

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		LogErrorf("Template render failed", err, "page", page)
	}
}
