package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// HostSettings Contains settings specific to SFTP and other features.
type HostSettings struct {
	DefaultPath string `json:"default_path"` // Папка, которая откроется первой
}

// Value to write to the database.
func (s HostSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan for reading from the database.
func (s *HostSettings) Scan(value interface{}) error {
	if value == nil {
		*s = HostSettings{DefaultPath: "/"}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		// Если база вернула строку (бывает в SQLite)
		sStr, ok := value.(string)
		if !ok {
			return errors.New("type assertion to []byte/string failed")
		}
		b = []byte(sStr)
	}
	return json.Unmarshal(b, s)
}

// Host host model.
type Host struct {
	ID        int          `json:"id"`
	UserID    int          `json:"user_id"`
	Name      string       `json:"name"`
	Address   string       `json:"address"`
	Username  string       `json:"username"`
	Port      int          `json:"port,string"`
	AuthType  string       `json:"auth_type"`
	Password  string       `json:"password,omitempty"`
	KeyID     *int         `json:"key_id,string"`
	Settings  HostSettings `json:"settings"`
	CreatedAt time.Time    `json:"created_at"`
}
