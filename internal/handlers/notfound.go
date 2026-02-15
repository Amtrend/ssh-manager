package handlers

import (
	"net/http"
	"ssh_manager/internal/utils"
)

// NotFoundHandler displays the page when 4xx errors occur.
func (h *Handlers) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	utils.RenderTemplate(w, "4xx.html", map[string]interface{}{
		"Title": "Page Not Found",
		"Code":  "404",
	}, r)
}
