package repository

import (
	"context"
	"ssh_manager/internal/models"
)

type UserRepository struct {
	DB DBTX
}

// GetByUsername gets user data by name.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	query := Rebind(`SELECT id, username, password_hash FROM users WHERE username = $1`)
	err := r.DB.QueryRowContext(ctx, query, username).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateUSername updates the username by its ID.
func (r *UserRepository) UpdateUSername(ctx context.Context, userID int, newUsername string) error {
	query := Rebind(`UPDATE users SET username = $1 WHERE id = $2`)
	_, err := r.DB.ExecContext(ctx, query, newUsername, userID)
	return err
}

// UpdatePassword updates the user's password by its ID.
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int, passwordHash string) error {
	query := Rebind(`UPDATE users SET password_hash = $1 WHERE id = $2`)
	_, err := r.DB.ExecContext(ctx, query, passwordHash, userID)
	return err
}
