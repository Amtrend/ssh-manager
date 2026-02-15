package handlers

import (
	"net/http"
	"ssh_manager/internal/utils"
)

// LogoutHandler logout handler.
func (h *Handlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := h.Store.Get(r, utils.SessionName)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	session.Values["authenticated"] = false
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Session save error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
