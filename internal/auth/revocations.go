package auth

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"

	appmiddleware "hnu-tickets/internal/platform/middleware"
)

// Revocations persists revoked token IDs until their original expiry.
type Revocations interface {
	IsRevoked(context.Context, uuid.UUID) (bool, error)
	Revoke(context.Context, uuid.UUID, time.Time) error
}

type PostgresRevocations struct{ pool *pgxpool.Pool }

func NewPostgresRevocations(pool *pgxpool.Pool) *PostgresRevocations {
	return &PostgresRevocations{pool: pool}
}

func (r *PostgresRevocations) IsRevoked(ctx context.Context, id uuid.UUID) (bool, error) {
	var revoked bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE id = $1 AND expires_at > now())`, id).Scan(&revoked)
	if err != nil {
		return false, fmt.Errorf("check token revocation: %w", err)
	}
	return revoked, nil
}

func (r *PostgresRevocations) Revoke(ctx context.Context, id uuid.UUID, expires time.Time) error {
	_, err := r.pool.Exec(ctx, `WITH expired AS (DELETE FROM revoked_tokens WHERE expires_at <= now())
		INSERT INTO revoked_tokens(id, expires_at) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, id, expires)
	if err != nil {
		return fmt.Errorf("revoke access token: %w", err)
	}
	return nil
}

// Logout invalidates only the bearer token used for this request.
func (s *Service) Logout(ctx context.Context, actor appmiddleware.Principal) error {
	return s.revocations.Revoke(ctx, actor.TokenID, actor.ExpiresAt)
}
