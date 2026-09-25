package comments

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/platform/httpx"
)

func TestMapHTTPError(t *testing.T) {
	mapped := MapHTTPError(ErrInvalidComment)
	var transportErr *httpx.Error
	require.ErrorAs(t, mapped, &transportErr)
	require.Equal(t, http.StatusBadRequest, transportErr.Status())
	require.Equal(t, codeInvalidComment, transportErr.Code())
}
