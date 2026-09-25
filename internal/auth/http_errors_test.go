package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/platform/httpx"
	"hnu-tickets/internal/users"
)

func TestMapHTTPError(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   httpx.Code
	}{
		{ErrInvalidPassword, http.StatusBadRequest, codeInvalidPassword},
		{ErrInvalidCredentials, http.StatusUnauthorized, codeInvalidCredentials},
		{users.ErrUserNotFound, http.StatusNotFound, "USER_NOT_FOUND"},
	}

	for _, test := range tests {
		mapped := mapHTTPError(test.err)
		var transportErr *httpx.Error
		require.ErrorAs(t, mapped, &transportErr)
		require.Equal(t, test.status, transportErr.Status())
		require.Equal(t, test.code, transportErr.Code())
	}
}
