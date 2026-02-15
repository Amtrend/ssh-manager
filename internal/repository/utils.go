package repository

import (
	"os"
	"regexp"
)

var rePostgresPlaceholder = regexp.MustCompile(`\$\d+`)

// Rebind replaces $N placeholders with ? if DB_TYPE=sqlite is set in the environment.
// This allows you to write Postgres-style queries and run them in SQLite.
func Rebind(query string) string {
	if os.Getenv("DB_TYPE") != "postgres" {
		return rePostgresPlaceholder.ReplaceAllString(query, "?")
	}
	return query
}
