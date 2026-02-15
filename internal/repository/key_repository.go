package repository

import (
	"context"
	"ssh_manager/internal/models"
)

type KeyRepository struct {
	DB DBTX
}

// GetAllByUserID gets all private keys of a user by his ID.
func (r *KeyRepository) GetAllByUserID(ctx context.Context, userID int) ([]models.Key, error) {
	query := Rebind(`SELECT id, name FROM keys WHERE user_id = $1 ORDER BY created_at DESC`)
	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.Key
	for rows.Next() {
		var k models.Key
		if err := rows.Scan(&k.ID, &k.Name); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// GetByID gets user key data by its ID.
func (r *KeyRepository) GetByID(ctx context.Context, keyID, userID int) (*models.Key, error) {
	k := &models.Key{}
	query := Rebind(`SELECT id, name, private_key FROM keys WHERE id = $1 AND user_id = $2`)
	err := r.DB.QueryRowContext(ctx, query, keyID, userID).Scan(&k.ID, &k.Name, &k.KeyData)
	return k, err
}

// Create creates a key.
func (r *KeyRepository) Create(ctx context.Context, key *models.Key) error {
	query := Rebind(`INSERT INTO keys (user_id, name, private_key) VALUES ($1, $2, $3) RETURNING id`)
	return r.DB.QueryRowContext(ctx, query, key.UserID, key.Name, key.KeyData).Scan(&key.ID)
}

// Update updates the key.
func (r *KeyRepository) Update(ctx context.Context, key *models.Key) error {
	query := Rebind(`UPDATE keys SET name = $1, private_key = $2 WHERE id = $3 AND user_id = $4`)
	_, err := r.DB.ExecContext(ctx, query, key.Name, key.KeyData, key.ID, key.UserID)
	return err
}

// Delete deletes a key by ID.
func (r *KeyRepository) Delete(ctx context.Context, keyID, userID int) error {
	query := Rebind(`DELETE FROM keys WHERE id = $1 AND user_id = $2`)
	_, err := r.DB.ExecContext(ctx, query, keyID, userID)
	return err
}
