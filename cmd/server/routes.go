package main

import (
	"net/http"
	"ssh_manager"
	"ssh_manager/internal/handlers"
	"ssh_manager/internal/middleware"
	"ssh_manager/internal/utils"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func SetupRoutes(h *handlers.Handlers, m *middleware.Middleware, store *sessions.CookieStore) *mux.Router {
	r := mux.NewRouter()

	// --- Global Middleware ---
	// ErrorHandlerMiddleware must be the very first to catch panic everywhere
	r.Use(middleware.ErrorHandlerMiddleware)
	// CSRFMiddleware prepares a token for all pages
	r.Use(middleware.CSRFMiddleware(store))

	// --- Statics ---
	staticHandler := http.FileServer(http.FS(ssh_manager.StaticFS))
	r.PathPrefix("/static/").Handler(staticHandler)

	// Service worker
	r.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		data, err := ssh_manager.StaticFS.ReadFile("static/js/sw.js")
		if err != nil {
			utils.LogErrorf("SW file not found in embed: %v", err)
			http.Error(w, "Not found", 404)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(data)
	})

	// --- Public routes ---
	r.HandleFunc("/login", h.LoginHandler).Methods("GET")
	r.HandleFunc("/login", h.LoginPostHandler).Methods("POST")

	// --- Protected routes ---
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(m.AuthMiddleware)
	protected.Use(middleware.CSRFValidationMiddleware)

	// Home and exit
	protected.HandleFunc("/", h.HomeHandler).Methods("GET")
	protected.HandleFunc("/logout", h.LogoutHandler).Methods("GET")

	// Profile
	profile := protected.PathPrefix("/profile").Subrouter()
	profile.HandleFunc("", h.ProfileHandler).Methods("GET")
	profile.HandleFunc("/update-username", h.UpdateUsernameHandler).Methods("POST")
	profile.HandleFunc("/update-password", h.UpdatePasswordHandler).Methods("POST")
	profile.HandleFunc("/push/subscribe", h.SubscribePushHandler).Methods("POST")
	profile.HandleFunc("/push/unsubscribe", h.UnsubscribePushHandler).Methods("POST")
	profile.HandleFunc("/push/reset", h.ResetNotificationsHandler).Methods("POST")

	// Keys
	keys := protected.PathPrefix("/keys").Subrouter()
	keys.HandleFunc("", h.KeysHandler).Methods("GET")
	keys.HandleFunc("/list", h.ListKeysHandler).Methods("GET")
	keys.HandleFunc("/add", h.AddKeyHandler).Methods("POST")
	keys.HandleFunc("/edit/{id:[0-9]+}", h.GetKeyDataHandler).Methods("GET")
	keys.HandleFunc("/edit/{id:[0-9]+}", h.EditKeyHandler).Methods("POST")
	keys.HandleFunc("/delete/{id:[0-9]+}", h.DeleteKeyHandler).Methods("POST")

	// Hosts
	hosts := protected.PathPrefix("/hosts").Subrouter()
	hosts.HandleFunc("/add", h.AddHostHandler).Methods("POST")
	hosts.HandleFunc("/data/{id:[0-9]+}", h.GetHostDataHandler).Methods("GET")
	hosts.HandleFunc("/edit/{id:[0-9]+}", h.EditHostHandler).Methods("POST")
	hosts.HandleFunc("/delete/{id:[0-9]+}", h.DeleteHostHandler).Methods("POST")

	// Websocket and termination
	protected.HandleFunc("/ws/ssh", h.SSHWebsocketHandler)
	protected.HandleFunc("/ssh/terminate", h.TerminateSessionHandler).Methods("POST")

	// SFTP
	sfpts := protected.PathPrefix("/sftp").Subrouter()
	sfpts.HandleFunc("/list", h.GetFilesHandler).Methods("GET")
	sfpts.HandleFunc("/download", h.DownloadFileHandler).Methods("GET")
	sfpts.HandleFunc("/download-zip", h.DownloadZipHandler).Methods("GET")
	sfpts.HandleFunc("/upload", h.UploadHandler).Methods("POST")

	// --- Processing 404 ---
	r.NotFoundHandler = http.HandlerFunc(h.NotFoundHandler)

	return r
}
