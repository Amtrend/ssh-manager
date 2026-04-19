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

// AddSubscription keeps the notification subscription.
func (r *UserRepository) AddSubscription(ctx context.Context, sub models.PushSubscription) error {
	query := Rebind(`INSERT INTO user_subscriptions (user_id, endpoint, p256dh, auth) VALUES ($1, $2, $3, $4) 
                     ON CONFLICT (endpoint) DO NOTHING`)
	_, err := r.DB.ExecContext(ctx, query, sub.UserID, sub.Endpoint, sub.P256dh, sub.Auth)
	return err
}

// RemoveSubscription deletes a subscription for a specific user.
func (r *UserRepository) RemoveSubscription(ctx context.Context, userID int) error {
	query := Rebind(`DELETE FROM user_subscriptions WHERE user_id = $1`)
	_, err := r.DB.ExecContext(ctx, query, userID)
	return err
}

// GetUserSubscriptions returns all active push subscriptions for a user.
func (r *UserRepository) GetUserSubscriptions(ctx context.Context, userID int) ([]models.PushSubscription, error) {
	var subs []models.PushSubscription
	query := Rebind(`SELECT user_id, endpoint, p256dh, auth FROM user_subscriptions WHERE user_id = $1`)

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s models.PushSubscription
		if err := rows.Scan(&s.UserID, &s.Endpoint, &s.P256dh, &s.Auth); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subs, nil
}

// ResetUnreadCount resets the counter of unread notifications
func (r *UserRepository) ResetUnreadCount(ctx context.Context, userID int) error {
	query := Rebind(`
		INSERT INTO user_notification_counts (user_id, unread_count)
		VALUES ($1, 0)
		ON CONFLICT (user_id) DO UPDATE SET unread_count = 0`)
	_, err := r.DB.ExecContext(ctx, query, userID)
	return err
}

// IncrementAndGetUnreadCount adds a value to the unread notification counter
func (r *UserRepository) IncrementAndGetUnreadCount(ctx context.Context, userID int) (int, error) {
	var count int
	query := Rebind(`
		INSERT INTO user_notification_counts (user_id, unread_count)
		VALUES ($1, 1)
		ON CONFLICT (user_id)
		DO UPDATE SET unread_count = user_notification_counts.unread_count + 1
		RETURNING unread_count`)
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}
