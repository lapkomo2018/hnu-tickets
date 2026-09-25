package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPasswordHash(t *testing.T) {
	hash, err := hashPassword("password123")
	require.NoError(t, err)
	require.NotEqual(t, "password123", hash)
	require.NoError(t, checkPassword(hash, "password123"))
	require.Error(t, checkPassword(hash, "wrong-password"))
}
