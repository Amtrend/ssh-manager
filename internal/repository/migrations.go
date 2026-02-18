package repository

import (
	"context"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// InitDB rolls out the pattern depending on the type of base.
func InitDB(db DBTX, dbType string) error {
	// SQL for Postgres (with SERIAL)
	pgSchema := `
	CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS keys (id SERIAL PRIMARY KEY, user_id INTEGER REFERENCES users(id), name TEXT NOT NULL, private_key TEXT NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS hosts (id SERIAL PRIMARY KEY, user_id INTEGER REFERENCES users(id), name TEXT NOT NULL, address TEXT NOT NULL, port INTEGER DEFAULT 22, username TEXT NOT NULL, auth_type TEXT DEFAULT 'key', password TEXT, key_id INTEGER REFERENCES keys(id), created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);
	`

	// SQL for SQLite (with AUTOINCREMENT and without TimeZone in the same syntax)
	sqliteSchema := `
	CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
    CREATE TABLE IF NOT EXISTS keys (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER REFERENCES users(id), name TEXT NOT NULL, private_key TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
    CREATE TABLE IF NOT EXISTS hosts (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER REFERENCES users(id), name TEXT NOT NULL, address TEXT NOT NULL, port INTEGER DEFAULT 22, username TEXT NOT NULL, auth_type TEXT DEFAULT 'key', password TEXT, key_id INTEGER REFERENCES keys(id), created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
	`

	schema := pgSchema
	if dbType == "sqlite" {
		schema = sqliteSchema
	}

	// We are creating tables
	for _, query := range strings.Split(schema, ";") {
		q := strings.TrimSpace(query)
		if q == "" {
			continue
		}
		if _, err := db.ExecContext(context.Background(), q); err != nil {
			return err
		}
	}
	log.Printf("Database schema initialized for %s", dbType)
	return nil
}

// EnsureAdminUser During initialization, it creates a user with a password in the application.
func EnsureAdminUser(db DBTX, dbType string, defaultUser, defaultPass string) {
	// Hashing the password
	hash, _ := bcrypt.GenerateFromPassword([]byte(defaultPass), bcrypt.DefaultCost)

	var query string
	if dbType == "postgres" {
		query = `INSERT INTO users (username, password_hash) VALUES ($1, $2) ON CONFLICT (username) DO NOTHING`
	} else {
		// SQLite requires a separate check, as older versions do not always support ON CONFLICT
		query = `INSERT OR IGNORE INTO users (username, password_hash) VALUES (?, ?)`
	}

	_, err := db.ExecContext(context.Background(), query, defaultUser, string(hash))
	if err != nil {
		log.Printf("Failed to ensure admin user: %v", err)
	} else {
		log.Printf("Initial user check done (User: %s)", defaultUser)
	}
}
