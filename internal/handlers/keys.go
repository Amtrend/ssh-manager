package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"ssh_manager/internal/encryption"
	"ssh_manager/internal/models"
	"ssh_manager/internal/utils"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// KeysHandler displaying a page with keys.
func (h *Handlers) KeysHandler(w http.ResponseWriter, r *http.Request) {
	sesion, _ := h.Store.Get(r, utils.SessionName)
	userID := sesion.Values[utils.UserIDKey].(int)

	keys, err := h.KeyRepo.GetAllByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("[ERROR] KeysHandler: %v", err)
		utils.SendJSONResponse(w, false, "Database error", 500)
		return
	}

	utils.RenderTemplate(w, "keys.html", map[string]interface{}{
		"Title":    "Keys",
		"Keys":     keys,
		"ShowMenu": true,
	}, r)
}

// ListKeysHandler returns the user's keys.
func (h *Handlers) ListKeysHandler(w http.ResponseWriter, r *http.Request) {
	sesion, _ := h.Store.Get(r, utils.SessionName)
	userID := sesion.Values[utils.UserIDKey].(int)

	keys, err := h.KeyRepo.GetAllByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("[ERROR] KeysHandler: %v", err)
		utils.SendJSONResponse(w, false, "Database error", 500)
		return
	}

	utils.SendJSONResponse(w, true, "Keys retrieved successfully", keys)
}

// GetKeyDataHandler returns the key.
func (h *Handlers) GetKeyDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	session, _ := h.Store.Get(r, utils.SessionName)
	userID := session.Values[utils.UserIDKey].(int)

	key, err := h.KeyRepo.GetByID(r.Context(), id, userID)
	if err != nil {
		utils.SendJSONResponse(w, false, "Mot found", nil)
		return
	}

	decrypted, _ := encryption.Decrypt(key.KeyData)
	key.KeyData = utils.FormatPartialKey(decrypted)

	// MASKING: Show only the edges
	// For example: "MIIEp... (masked) ...short_hash"
	if len(decrypted) > 32 {
		key.KeyData = decrypted[:15] + "...... (private key hidden for security) ......" + decrypted[len(decrypted)-15:]
	}

	utils.SendJSONResponse(w, true, "", key)
}

// AddKeyHandler saves the new key.
func (h *Handlers) AddKeyHandler(w http.ResponseWriter, r *http.Request) {
	var key models.Key
	err := json.NewDecoder(r.Body).Decode(&key)
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid request data", nil)
		return
	}

	trimmedKeyData := strings.TrimSpace(key.KeyData)
	if trimmedKeyData == "" {
		utils.SendJSONResponse(w, false, "Private key cannot be empty", nil)
		return
	}

	session, _ := h.Store.Get(r, utils.SessionName)
	key.UserID = session.Values[utils.UserIDKey].(int)

	// Encrypting a private key
	encrypted, _ := encryption.Encrypt(trimmedKeyData)
	key.KeyData = encrypted

	if err := h.KeyRepo.Create(r.Context(), &key); err != nil {
		utils.SendJSONResponse(w, false, "DB Error", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Key added successfully", key)
}

// EditKeyHandler edits an existing key.
func (h *Handlers) EditKeyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid key ID", nil)
		return
	}

	var key models.Key
	if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
		utils.SendJSONResponse(w, false, "Invalid request data", nil)
		return
	}

	session, _ := h.Store.Get(r, utils.SessionName)
	userID := session.Values[utils.UserIDKey].(int)

	oldKey, err := h.KeyRepo.GetByID(r.Context(), id, userID)
	if err != nil {
		log.Printf("[ERROR] EditKey: Key %d not found for user %d: %v", id, userID, err)
		utils.SendJSONResponse(w, false, "Key not found", nil)
		return
	}

	updatedKeyData := oldKey.KeyData
	trimmedKeyData := strings.TrimSpace(key.KeyData)

	if !strings.Contains(trimmedKeyData, "...") && trimmedKeyData != "" {
		encrypted, err := encryption.Encrypt(trimmedKeyData)
		if err != nil {
			utils.SendJSONResponse(w, false, "Encryption error", nil)
			return
		}
		updatedKeyData = encrypted
	} else if trimmedKeyData == "" {
		utils.SendJSONResponse(w, false, "Private key cannot be empty", nil)
		return
	}

	keyToUpdate := &models.Key{
		ID:      id,
		UserID:  userID,
		Name:    key.Name,
		KeyData: updatedKeyData,
	}

	if err := h.KeyRepo.Update(r.Context(), keyToUpdate); err != nil {
		log.Printf("[ERROR] KeyRepo.Update: %v", err)
		utils.SendJSONResponse(w, false, "Database error", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Key updated successfully", map[string]interface{}{
		"id":   id,
		"name": key.Name,
	})
}

// DeleteKeyHandler deletes the key.
func (h *Handlers) DeleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid key ID", nil)
		return
	}

	session, _ := h.Store.Get(r, utils.SessionName)
	userID := session.Values[utils.UserIDKey].(int)

	if err := h.KeyRepo.Delete(r.Context(), id, userID); err != nil {
		log.Printf("[ERROR] KeyRepo.Delete (ID: %d, User: %d): %v", id, userID, err)
		utils.SendJSONResponse(w, false, "Database error", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Key deleted successfully", map[string]interface{}{
		"id": id,
	})
}
