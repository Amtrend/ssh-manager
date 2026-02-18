package middleware

import (
	"fmt"
	"net/http"
	"ssh_manager/internal/utils"
)

// ErrorHandlerMiddleware handles errors and displays custom pages.
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				var err error
				switch t := rvr.(type) {
				case error:
					err = t
				default:
					err = fmt.Errorf("%v", t)
				}

				utils.LogErrorf("Panic recovered", err,
					"method", r.Method,
					"path", r.URL.Path,
				)

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
