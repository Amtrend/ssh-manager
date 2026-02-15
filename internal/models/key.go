package models

import "time"

// Key represents a private key model.
type Key struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	KeyData   string    `json:"key_data"`
	CreatedAt time.Time `json:"created_at"`
}
