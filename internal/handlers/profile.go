package handlers

import (
	"encoding/json"
	"net/http"
	"ssh_manager/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

// ProfileHandler displays profile page.
func (h *Handlers) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, err := h.Store.Get(r, utils.SessionName)
	if err != nil {
		utils.SendJSONResponse(w, false, "Session error", nil)
		return
	}

	username, _ := session.Values[utils.UsernameKey].(string)
	utils.RenderTemplate(w, "profile.html", map[string]interface{}{
		"Title":    "Profile",
		"Username": username,
		"ShowMenu": true,
	}, r)
}

// UpdateUsernameHandler updates username.
func (h *Handlers) UpdateUsernameHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Username string `json:"username"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid request data", nil)
		return
	}

	session, _ := h.Store.Get(r, utils.SessionName)
	userID := session.Values[utils.UserIDKey].(int)

	if err := h.UserRepo.UpdateUSername(r.Context(), userID, requestData.Username); err != nil {
		utils.SendJSONResponse(w, false, "Database error", nil)
		return
	}

	session.Values[utils.UsernameKey] = requestData.Username
	err = session.Save(r, w)
	if err != nil {
		utils.SendJSONResponse(w, false, "Session save error", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Username updated successfully", nil)
}

// UpdatePasswordHandler updates user password.
func (h *Handlers) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		NewPassword string `json:"new_password"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid request data", nil)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestData.NewPassword), bcrypt.DefaultCost)

	session, _ := h.Store.Get(r, utils.SessionName)
	userID := session.Values[utils.UserIDKey].(int)

	if err != nil {
		utils.SendJSONResponse(w, false, "Password hashing error", nil)
		return
	}

	if err := h.UserRepo.UpdatePassword(r.Context(), userID, string(hashedPassword)); err != nil {
		utils.SendJSONResponse(w, false, "Database error", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Password updated successfully", nil)
}
