package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ssh_manager/internal/models"
	"ssh_manager/internal/utils"

	"github.com/gorilla/mux"
)

// GetHostDataHandler returns data for a specific host for editing.
func (h *Handlers) GetHostDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid host ID", nil)
		return
	}

	session, _ := h.Store.Get(r, utils.SessionName)
	userID := session.Values[utils.UserIDKey].(int)

	host, err := h.HostRepo.GetByID(r.Context(), id, userID)
	if err != nil {
		utils.SendJSONResponse(w, false, "Host not found", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Host data retrieved successfully", host)
}

// AddHostHandler adds a new host.
func (h *Handlers) AddHostHandler(w http.ResponseWriter, r *http.Request) {
	var host models.Host
	err := json.NewDecoder(r.Body).Decode(&host)
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid request data"+err.Error(), nil)
		return
	}

	// Adding a host to the database
	session, _ := h.Store.Get(r, utils.SessionName)
	host.UserID = session.Values[utils.UserIDKey].(int)

	if err := h.HostRepo.Create(r.Context(), &host); err != nil {
		utils.SendJSONResponse(w, false, "Failed to add host", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Host added successfully", host)
}

// EditHostHandler updates an existing host.
func (h *Handlers) EditHostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid host ID", nil)
		return
	}

	var host models.Host
	err = json.NewDecoder(r.Body).Decode(&host)
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid request data", nil)
		return
	}
	host.ID = id

	session, _ := h.Store.Get(r, utils.SessionName)
	host.UserID = session.Values[utils.UserIDKey].(int)

	if err := h.HostRepo.Update(r.Context(), &host); err != nil {
		utils.SendJSONResponse(w, false, "Failed to update host", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Host updated successfully", nil)
}

// DeleteHostHandler deletes host.
func (h *Handlers) DeleteHostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid host ID", nil)
		return
	}

	session, _ := h.Store.Get(r, utils.SessionName)
	userID := session.Values[utils.UserIDKey].(int)

	err = h.HostRepo.Delete(r.Context(), id, userID)
	if err != nil {
		utils.SendJSONResponse(w, false, "Failed to delete host", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Host deleted successfully", map[string]interface{}{
		"id": id,
	})
}
