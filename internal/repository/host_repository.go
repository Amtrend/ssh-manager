package repository

import (
	"context"
	"ssh_manager/internal/models"
)

type HostRepository struct {
	DB DBTX
}

// GetByUserID gets data from all user hosts by their ID.
func (r *HostRepository) GetByUserID(ctx context.Context, userID int) ([]models.Host, error) {
	query := Rebind(`SELECT id, name, address, port, username, key_id FROM hosts WHERE user_id = $1`)
	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []models.Host
	for rows.Next() {
		var h models.Host
		err := rows.Scan(
			&h.ID,
			&h.Name,
			&h.Address,
			&h.Port,
			&h.Username,
			&h.KeyID,
		)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

// GetByID gets host data by user ID and host ID.
func (r *HostRepository) GetByID(ctx context.Context, hostID, userID int) (*models.Host, error) {
	h := &models.Host{}
	query := Rebind(`SELECT id, name, address, port, username, key_id FROM hosts WHERE id = $1 AND user_id = $2`)
	err := r.DB.QueryRowContext(ctx, query, hostID, userID).Scan(&h.ID, &h.Name, &h.Address, &h.Port, &h.Username, &h.KeyID)
	return h, err
}

// Create will create a host.
func (r *HostRepository) Create(ctx context.Context, h *models.Host) error {
	query := Rebind(`INSERT INTO hosts (user_id, name, address, port, username, key_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`)
	return r.DB.QueryRowContext(ctx, query, h.UserID, h.Name, h.Address, h.Port, h.Username, h.KeyID).Scan(&h.ID)
}

// Update updates host data.
func (r *HostRepository) Update(ctx context.Context, h *models.Host) error {
	query := Rebind(`UPDATE hosts SET name = $1, address = $2, port = $3, username = $4, key_id = $5 WHERE id = $6 AND user_id = $7`)
	_, err := r.DB.ExecContext(ctx, query, h.Name, h.Address, h.Port, h.Username, h.KeyID, h.ID, h.UserID)
	return err
}

// Delete deletes a host by its ID.
func (r *HostRepository) Delete(ctx context.Context, hostID, userID int) error {
	query := Rebind(`DELETE FROM hosts WHERE id = $1 AND user_id = $2`)
	_, err := r.DB.ExecContext(ctx, query, hostID, userID)
	return err
}
