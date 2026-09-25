package users

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/platform/httpx"
)

func TestMapHTTPError(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   httpx.Code
	}{
		{ErrUserNotFound, http.StatusNotFound, codeUserNotFound},
		{ErrEmailAlreadyExists, http.StatusConflict, codeEmailAlreadyExists},
		{ErrInvalidName, http.StatusBadRequest, codeInvalidName},
		{ErrInvalidEmail, http.StatusBadRequest, codeInvalidEmail},
		{ErrInvalidRole, http.StatusBadRequest, codeInvalidRole},
		{ErrAdminRequired, http.StatusForbidden, codeForbidden},
	}

	for _, test := range tests {
		mapped := MapHTTPError(test.err)
		var transportErr *httpx.Error
		require.ErrorAs(t, mapped, &transportErr)
		require.Equal(t, test.status, transportErr.Status())
		require.Equal(t, test.code, transportErr.Code())
	}
}
