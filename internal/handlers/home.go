package handlers

import (
	"log"
	"net/http"
	"ssh_manager/internal/utils"
)

// HomeHandler displays the main page with hosts.
func (h *Handlers) HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Loading hosts from the database
	session, _ := h.Store.Get(r, utils.SessionName)
	userID := session.Values[utils.UserIDKey].(int)

	hosts, err := h.HostRepo.GetByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("[ERROR] HomeHandler - GetByUserID (userID: %d): %v", userID, err)
		utils.SendJSONResponse(w, false, "Database error", nil)
		return
	}

	// active connections to SSH
	activeIDs := h.SSHService.GetActiveHostIDs(userID)

	utils.RenderTemplate(w, "home.html", map[string]interface{}{
		"Title":     "Home",
		"ShowMenu":  true,
		"Hosts":     hosts,
		"ActiveIDs": activeIDs,
	}, r)
}
