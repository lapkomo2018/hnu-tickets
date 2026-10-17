package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func setRequired(t *testing.T) {
	t.Helper()
	for _, key := range []string{"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_SSLMODE", "ADMIN_EMAIL", "ADMIN_PASSWORD"} {
		t.Setenv(key, "")
	}
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("POSTGRES_USER", "tickets")
	t.Setenv("POSTGRES_PASSWORD", "tickets")
	t.Setenv("POSTGRES_DB", "tickets")
}

func TestLoadBuildsDatabaseURLWithDefaults(t *testing.T) {
	setRequired(t)

	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, "postgres://tickets:tickets@localhost:5432/tickets?sslmode=disable", cfg.Database.URL())
}

func TestLoadBuildsDatabaseURLFromEnvironment(t *testing.T) {
	setRequired(t)
	t.Setenv("POSTGRES_HOST", "db.internal")
	t.Setenv("POSTGRES_PORT", "6543")
	t.Setenv("POSTGRES_PASSWORD", "p@ss/word")
	t.Setenv("POSTGRES_SSLMODE", "require")

	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, "postgres://tickets:p%40ss%2Fword@db.internal:6543/tickets?sslmode=require", cfg.Database.URL())
}

func TestLoadRequiresDatabaseCredentials(t *testing.T) {
	setRequired(t)
	t.Setenv("POSTGRES_PASSWORD", "")

	_, err := Load()

	require.EqualError(t, err, "POSTGRES_PASSWORD is required")
}
