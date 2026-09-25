package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTokenManagerRoundTrip(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)
	userID := uuid.New()

	raw, err := manager.Issue(userID, "ADMIN")
	require.NoError(t, err)

	principal, err := manager.Parse(raw)
	require.NoError(t, err)
	require.Equal(t, userID, principal.UserID)
	require.Equal(t, "ADMIN", principal.Role)
}

func TestTokenManagerRejectsInvalidToken(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)
	_, err := manager.Parse("not-a-token")
	require.Error(t, err)
}
