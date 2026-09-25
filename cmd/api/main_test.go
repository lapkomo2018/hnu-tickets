package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/platform/config"
)

func TestHealthEndpoint(t *testing.T) {
	router := newRouter(nil, testConfig())
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"status":"ok"}`, response.Body.String())
}

func TestProtectedEndpointRequiresToken(t *testing.T) {
	router := newRouter(nil, testConfig())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tickets", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.JSONEq(t, `{"error":{"code":"UNAUTHORIZED","message":"a bearer token is required"}}`, response.Body.String())
}

func TestRegisterRejectsInvalidJSON(t *testing.T) {
	router := newRouter(nil, testConfig())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"name":`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{"error":{"code":"INVALID_JSON","message":"request body is not valid JSON"}}`, response.Body.String())
}

func testConfig() config.Config {
	return config.Config{
		JWT:  config.JWTConfig{Secret: "test-secret", TTL: time.Hour},
		CORS: config.CORSConfig{AllowedOrigins: []string{"http://localhost:5173"}},
	}
}
