package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"ssh_manager/internal/models"
	"ssh_manager/internal/utils"
	"time"

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
		"Title":          "Profile",
		"Username":       username,
		"ShowMenu":       true,
		"VapidPublicKey": utils.GetEnv("VAPID_PUBLIC_KEY", ""),
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

// SubscribePushHandler subscribes the user to push notifications.
func (h *Handlers) SubscribePushHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
		P256dh   string `json:"p256dh"`
		Auth     string `json:"auth"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendJSONResponse(w, false, "Invalid request data", nil)
		return
	}

	session, _ := h.Store.Get(r, utils.SessionName)
	userID, ok := session.Values[utils.UserIDKey].(int)
	if !ok {
		utils.SendJSONResponse(w, false, "Unauthorized", nil)
		return
	}

	sub := models.PushSubscription{
		UserID:   userID,
		Endpoint: req.Endpoint,
		P256dh:   req.P256dh,
		Auth:     req.Auth,
	}

	if err := h.UserRepo.AddSubscription(r.Context(), sub); err != nil {
		utils.LogErrorf("Error streaming file", err)
		utils.SendJSONResponse(w, false, "Database error", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Subscribed successfully", nil)
}

// UnsubscribePushHandler unsubscribe from push notifications.
func (h *Handlers) UnsubscribePushHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, utils.SessionName)
	userID, ok := session.Values[utils.UserIDKey].(int)
	if !ok {
		utils.SendJSONResponse(w, false, "Unauthorized", nil)
		return
	}

	if err := h.UserRepo.RemoveSubscription(r.Context(), userID); err != nil {
		utils.SendJSONResponse(w, false, "Database error", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Unsubscribed successfully", nil)
}

// ResetNotificationsHandler resetting the notification counter
func (h *Handlers) ResetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, utils.SessionName)
	userID, ok := session.Values[utils.UserIDKey].(int)
	if !ok {
		utils.SendJSONResponse(w, false, "Unauthorized", nil)
		return
	}

	// call reset in db
	ctx, calcel := context.WithTimeout(r.Context(), 5*time.Second)
	defer calcel()

	if err := h.UserRepo.ResetUnreadCount(ctx, userID); err != nil {
		utils.LogErrorf("Failed to reset notifications", err, "userID", userID)
		utils.SendJSONResponse(w, false, "Database error", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Notifications reset successfully", nil)
}
