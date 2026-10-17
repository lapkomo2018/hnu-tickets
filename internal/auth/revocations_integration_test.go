package auth

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestPostgresRevocationsPersistAndExpire(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is required for isolated PostgreSQL integration checks")
	}
	pool, err := pgxpool.New(t.Context(), dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	schema := "auth_test_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	quoted := pgx.Identifier{schema}.Sanitize()
	_, err = pool.Exec(t.Context(), "CREATE SCHEMA "+quoted)
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, dropErr := pool.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE")
		require.NoError(t, dropErr)
	})
	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	config.ConnConfig.RuntimeParams["search_path"] = schema
	first, err := pgxpool.NewWithConfig(t.Context(), config)
	require.NoError(t, err)
	t.Cleanup(first.Close)
	sql, err := os.ReadFile("../../migrations/000006_create_revoked_tokens.up.sql")
	require.NoError(t, err)
	_, err = first.Exec(t.Context(), string(sql))
	require.NoError(t, err)
	id := uuid.New()
	repository := NewPostgresRevocations(first)
	blocked, err := repository.IsRevoked(t.Context(), id)
	require.NoError(t, err)
	require.False(t, blocked)
	expires := time.Now().Add(time.Hour)
	require.NoError(t, repository.Revoke(t.Context(), id, expires))
	require.NoError(t, repository.Revoke(t.Context(), id, expires))
	second, err := pgxpool.NewWithConfig(t.Context(), config.Copy())
	require.NoError(t, err)
	t.Cleanup(second.Close)
	blocked, err = NewPostgresRevocations(second).IsRevoked(t.Context(), id)
	require.NoError(t, err)
	require.True(t, blocked)
	expiredID := uuid.New()
	require.NoError(t, repository.Revoke(t.Context(), expiredID, time.Now().Add(-time.Hour)))
	blocked, err = repository.IsRevoked(t.Context(), expiredID)
	require.NoError(t, err)
	require.False(t, blocked)
	require.NoError(t, repository.Revoke(t.Context(), uuid.New(), expires))
	var count int
	err = first.QueryRow(t.Context(), "SELECT count(*) FROM revoked_tokens WHERE id=$1", expiredID).Scan(&count)
	require.NoError(t, err)
	require.Zero(t, count)
}
