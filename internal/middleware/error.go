package middleware

import (
	"net/http"
	"ssh_manager/internal/utils"
)

// ErrorHandlerMiddleware handles errors and displays custom pages.
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				utils.LogErrorf("Panic: %v", err)

				// If this is an attempt to establish a WebSocket, simply close the connection.
				if r.Header.Get("Upgrade") == "websocket" {
					return
				}

				w.WriteHeader(http.StatusInternalServerError)
				utils.RenderTemplate(w, "5xx.html", map[string]interface{}{"Error": "Internal Error"}, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
