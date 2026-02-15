package middleware

import (
	"net/http"
	"ssh_manager/internal/utils"

	"github.com/gorilla/sessions"
)

type Middleware struct {
	Store *sessions.CookieStore
}

// AuthMiddleware checks whether the user is authorized.
func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Receiving a session
		session, err := m.Store.Get(r, utils.SessionName)
		if err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// Checking if the "authenticated" flag exists
		if auth, ok := session.Values[utils.IsAuthenticated].(bool); !ok || !auth {
			// If the user is not authorized, we redirect to the login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// If authorization is successful, we call the next handler
		next.ServeHTTP(w, r)
	})
}
