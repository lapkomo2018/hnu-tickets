package httpx

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorMiddlewareMapsTransportAndMappedErrors(t *testing.T) {
	mappedErr := errors.New("mapped")
	tests := []struct {
		name   string
		err    error
		mapper ErrorMapper
		status int
		code   Code
	}{
		{name: "transport", err: NewError(http.StatusBadRequest, CodeInvalidJSON, "invalid"), status: http.StatusBadRequest, code: CodeInvalidJSON},
		{name: "mapped", err: mappedErr, mapper: func(error) error {
			return Conflict("MAPPED", "mapped")
		}, status: http.StatusConflict, code: "MAPPED"},
		{name: "unknown", err: errors.New("failure"), status: http.StatusInternalServerError, code: CodeInternalError},
	}

	middleware := newTestErrorMiddleware()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			response := httptest.NewRecorder()
			wrap := middleware.Handle
			if test.mapper != nil {
				wrap = middleware.HandleWith(test.mapper)
			}
			handler := wrap(func(http.ResponseWriter, *http.Request) error {
				return test.err
			})

			handler.ServeHTTP(response, request)

			require.Equal(t, test.status, response.Code)
			require.Contains(t, response.Body.String(), `"code":"`+string(test.code)+`"`)
		})
	}
}

func TestErrorMiddlewareRecoversPanics(t *testing.T) {
	middleware := newTestErrorMiddleware()
	handler := middleware.Recoverer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), `"code":"INTERNAL_ERROR"`)
}

func newTestErrorMiddleware() *ErrorMiddleware {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewErrorMiddleware(logger)
}
