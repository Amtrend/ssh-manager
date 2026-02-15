package handlers

import (
	"ssh_manager/internal/repository"
	"ssh_manager/internal/services"

	"github.com/gorilla/sessions"
)

// Handlers contains common dependencies for all handlers.
type Handlers struct {
	UserRepo   *repository.UserRepository
	KeyRepo    *repository.KeyRepository
	HostRepo   *repository.HostRepository
	Store      *sessions.CookieStore
	SSHService *services.SSHService
}
