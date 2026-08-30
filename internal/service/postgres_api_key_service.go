package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"logmesh/internal/model"
)

type PostgresAPIKeyService struct {
	pool *pgxpool.Pool
}

func NewPostgresAPIKeyService(pool *pgxpool.Pool) *PostgresAPIKeyService {
	return &PostgresAPIKeyService{pool: pool}
}

func (s *PostgresAPIKeyService) Create(ctx context.Context, name string) (model.APIKey, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Untitled key"
	}

	plaintext, err := generateAPIKey()
	if err != nil {
		return model.APIKey{}, err
	}

	record := model.APIKey{
		ID:           uuid.NewString(),
		Name:         name,
		Prefix:       keyPrefix(plaintext),
		CreatedAt:    time.Now().UTC(),
		PlaintextKey: plaintext,
	}

	_, err = s.pool.Exec(ctx, `
INSERT INTO api_keys (id, name, prefix, key_hash, created_at)
VALUES ($1, $2, $3, $4, $5)
`, record.ID, record.Name, record.Prefix, hashAPIKey(plaintext), record.CreatedAt)
	if err != nil {
		return model.APIKey{}, err
	}

	return record, nil
}

func (s *PostgresAPIKeyService) List(ctx context.Context) ([]model.APIKey, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, name, prefix, created_at, last_used_at, revoked_at
FROM api_keys
ORDER BY created_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]model.APIKey, 0)
	for rows.Next() {
		var key model.APIKey
		var lastUsedAt pgtype.Timestamptz
		var revokedAt pgtype.Timestamptz
		if err := rows.Scan(&key.ID, &key.Name, &key.Prefix, &key.CreatedAt, &lastUsedAt, &revokedAt); err != nil {
			return nil, err
		}
		key.LastUsedAt = timestamptzPtr(lastUsedAt)
		key.RevokedAt = timestamptzPtr(revokedAt)
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (s *PostgresAPIKeyService) Revoke(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE api_keys SET revoked_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

func (s *PostgresAPIKeyService) Authenticate(ctx context.Context, plaintext string) (model.APIKey, bool) {
	hash := hashAPIKey(plaintext)

	var key model.APIKey
	var lastUsedAt pgtype.Timestamptz
	var revokedAt pgtype.Timestamptz
	err := s.pool.QueryRow(ctx, `
UPDATE api_keys
SET last_used_at = NOW()
WHERE key_hash = $1 AND revoked_at IS NULL
RETURNING id, name, prefix, created_at, last_used_at, revoked_at
`, hash).Scan(&key.ID, &key.Name, &key.Prefix, &key.CreatedAt, &lastUsedAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.APIKey{}, false
	}
	if err != nil {
		return model.APIKey{}, false
	}
	key.LastUsedAt = timestamptzPtr(lastUsedAt)
	key.RevokedAt = timestamptzPtr(revokedAt)
	return key, true
}

var _ APIKeyService = (*PostgresAPIKeyService)(nil)

func timestamptzPtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}
