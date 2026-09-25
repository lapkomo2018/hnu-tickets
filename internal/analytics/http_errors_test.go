package analytics

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
		{ErrAdminRequired, http.StatusForbidden, codeForbidden},
		{ErrInvalidPeriod, http.StatusBadRequest, codeInvalidPeriod},
	}

	for _, test := range tests {
		mapped := mapHTTPError(test.err)
		var transportErr *httpx.Error
		require.ErrorAs(t, mapped, &transportErr)
		require.Equal(t, test.status, transportErr.Status())
		require.Equal(t, test.code, transportErr.Code())
	}
}
