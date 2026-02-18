package models

import "time"

// Host host model.
type Host struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Username  string    `json:"username"`
	Port      int       `json:"port,string"`
	AuthType  string    `json:"auth_type"`
	Password  string    `json:"password,omitempty"`
	KeyID     *int      `json:"key_id,string"`
	CreatedAt time.Time `json:"created_at"`
}
