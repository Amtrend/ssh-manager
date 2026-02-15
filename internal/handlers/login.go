package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"ssh_manager/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

// LoginHandler displays the login page.
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Checking if the user is authorized
	session, _ := h.Store.Get(r, utils.SessionName)
	if auth, ok := session.Values[utils.IsAuthenticated].(bool); ok && auth {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Displaying the login page
	utils.RenderTemplate(w, "login.html", map[string]interface{}{
		"Title": "Login",
	}, r)
}

// LoginPostHandler authenticates and authorizes the user.
func (h *Handlers) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid request data", nil)
		return
	}

	user, err := h.UserRepo.GetByUsername(context.Background(), loginData.Username)
	if err != nil {
		log.Printf("[ERROR] LoginPostHandler - GetByUsername (%s): %v", loginData.Username, err)
		utils.SendJSONResponse(w, false, "Invalid username or password", nil)
		return
	}

	// Password comparison
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginData.Password))
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid username or password", nil)
		return
	}

	// Creating a session
	session, err := h.Store.Get(r, utils.SessionName)
	if err != nil {
		utils.SendJSONResponse(w, false, "Session error. Please try again later.", nil)
		return
	}
	session.Values[utils.IsAuthenticated] = true
	session.Values[utils.UsernameKey] = user.Username
	session.Values[utils.UserIDKey] = user.ID
	err = session.Save(r, w)
	if err != nil {
		utils.SendJSONResponse(w, false, "Session save error. Please try again later.", nil)
		return
	}

	utils.SendJSONResponse(w, true, "Login succesful", session.Values["authenticated"])
}
