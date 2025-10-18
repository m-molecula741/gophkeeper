package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/m-molecula741/gophkeeper/internal/domain"
)

var (
	ErrSecretNotFound = errors.New("secret not found")
)

type secretRepo struct {
	db *sql.DB
}

func NewSecretRepository(client *PostgresClient) *secretRepo {
	return &secretRepo{db: client.DB()}
}

func (r *secretRepo) Create(ctx context.Context, secret *domain.Secret) error {
	query := `
		INSERT INTO secrets (id, user_id, type, metadata, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		secret.ID,
		secret.UserID,
		secret.Type,
		secret.Metadata,
		secret.Data,
		secret.CreatedAt,
		secret.UpdatedAt,
	)
	return err
}

func (r *secretRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, metadata, data, created_at, updated_at
		FROM secrets
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query secrets: %w", err)
	}
	defer rows.Close()

	var secrets []*domain.Secret
	for rows.Next() {
		var s domain.Secret
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.Type,
			&s.Metadata,
			&s.Data,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}
		s.CreatedAt = createdAt
		s.UpdatedAt = updatedAt
		secrets = append(secrets, &s)
	}
	return secrets, nil
}

func (r *secretRepo) GetByID(ctx context.Context, id, userID string) (*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, metadata, data, created_at, updated_at
		FROM secrets
		WHERE id = $1 AND user_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, id, userID)

	var s domain.Secret
	var createdAt, updatedAt time.Time
	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.Type,
		&s.Metadata,
		&s.Data,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSecretNotFound
		}
		return nil, fmt.Errorf("failed to scan secret by ID: %w", err)
	}
	s.CreatedAt = createdAt
	s.UpdatedAt = updatedAt
	return &s, nil
}

func (r *secretRepo) Update(ctx context.Context, secret *domain.Secret) error {
	query := `
		UPDATE secrets
		SET type = $1, metadata = $2, data = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6
	`
	res, err := r.db.ExecContext(ctx, query,
		secret.Type,
		secret.Metadata,
		secret.Data,
		secret.UpdatedAt,
		secret.ID,
		secret.UserID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrSecretNotFound
	}
	return nil
}

func (r *secretRepo) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM secrets WHERE id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrSecretNotFound
	}
	return nil
}