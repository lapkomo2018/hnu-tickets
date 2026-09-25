package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/platform/config"
	"hnu-tickets/internal/platform/httpx"
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
	requireAPIError(t, response, httpx.CodeUnauthorized, "a bearer token is required")
}

func TestRegisterRejectsInvalidJSON(t *testing.T) {
	router := newRouter(nil, testConfig())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"name":`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	requireAPIError(t, response, httpx.CodeInvalidJSON, "request body is not valid JSON")
}

func TestRegisterRejectsMultipleJSONValues(t *testing.T) {
	router := newRouter(nil, testConfig())
	body := `{"name":"User","email":"user@example.com","password":"password123"} {}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	requireAPIError(t, response, httpx.CodeInvalidJSON, "request body must contain one JSON value")
}

func TestUnknownRouteUsesStructuredError(t *testing.T) {
	router := newRouter(nil, testConfig())
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNotFound, response.Code)
	requireAPIError(t, response, httpx.CodeRouteNotFound, "route not found")
}

func TestUnsupportedMethodUsesStructuredError(t *testing.T) {
	router := newRouter(nil, testConfig())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusMethodNotAllowed, response.Code)
	requireAPIError(t, response, httpx.CodeMethodNotAllowed, "method not allowed")
}

func testConfig() config.Config {
	return config.Config{
		JWT:  config.JWTConfig{Secret: "test-secret", TTL: time.Hour},
		CORS: config.CORSConfig{AllowedOrigins: []string{"http://localhost:5173"}},
	}
}

func requireAPIError(t *testing.T, response *httptest.ResponseRecorder, code httpx.Code, message string) {
	t.Helper()
	var envelope httpx.ErrorEnvelope
	require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))
	require.Equal(t, code, envelope.Error.Code)
	require.Equal(t, message, envelope.Error.Message)
}
