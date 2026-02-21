package middleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"ssh_manager/internal/utils"
	"strings"

	"github.com/gorilla/sessions"
)

// CSRFMiddleware adds a CSRF token to the session.
func CSRFMiddleware(store *sessions.CookieStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, utils.SessionName)
			if err != nil {
				utils.SendJSONResponse(w, false, "Session error", nil)
				return
			}

			// Generate a CSRF token if it is not already in the session
			if session.Values["csrf_token"] == nil {
				csrfToken := generateCSRFToken()
				session.Values["csrf_token"] = csrfToken
				err = session.Save(r, w)
				if err != nil {
					utils.SendJSONResponse(w, false, "Session save error", nil)
					return
				}
			}

			// Passing a CSRF token to the request context
			r = r.WithContext(context.WithValue(r.Context(), "csrf_token", session.Values["csrf_token"]))

			next.ServeHTTP(w, r)
		})
	}
}

// CSRFValidationMiddleware checks the CSRF token.
func CSRFValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			expectedToken, ok := r.Context().Value("csrf_token").(string)
			if !ok {
				utils.SendJSONResponse(w, false, "CSRF token not found", nil)
				return
			}

			var actualToken string
			contentType := r.Header.Get("Content-Type")

			// If it's JSON.
			if strings.Contains(contentType, "application/json") {
				body, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(body)) // Returning the body to the handler.

				var requestData map[string]interface{}
				if err := json.Unmarshal(body, &requestData); err == nil {
					actualToken, _ = requestData["csrf_token"].(string)
				}

				// If it is a file upload or a regular form.
			} else if strings.Contains(contentType, "multipart/form-data") ||
				strings.Contains(contentType, "application/x-www-form-urlencoded") {

				// For multipart, we parse the form to get the value.
				if err := r.ParseMultipartForm(32 << 20); err == nil {
					actualToken = r.FormValue("csrf_token")
				}
			}

			if actualToken == "" || actualToken != expectedToken {
				utils.SendJSONResponse(w, false, "Invalid CSRF token", nil)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// generateCSRFToken generating a random CSRF token.
func generateCSRFToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
