package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/platform/httpx"
	"hnu-tickets/internal/users"
)

type memoryRevocations struct {
	entries map[uuid.UUID]time.Time
	failure error
}

func (m *memoryRevocations) IsRevoked(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.entries[id]
	return ok, m.failure
}
func (m *memoryRevocations) Revoke(_ context.Context, id uuid.UUID, expires time.Time) error {
	if m.failure != nil {
		return m.failure
	}
	m.entries[id] = expires
	return nil
}
func (r *memoryUsers) GetByID(_ context.Context, id uuid.UUID) (*users.User, error) {
	for _, u := range r.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, users.ErrUserNotFound
}

func TestSessionRoleRefreshLogoutAndStorageFailure(t *testing.T) {
	userID := uuid.New()
	accounts := newMemoryUsers(users.User{ID: userID, Name: "User", Email: "user@example.com", Role: users.RoleAdmin})
	revoked := &memoryRevocations{entries: make(map[uuid.UUID]time.Time)}
	tokens := NewTokenManager("session-test-key", time.Hour)
	service := NewService(accounts, tokens, revoked)
	errorsMiddleware := httpx.NewErrorMiddleware(slog.New(slog.NewTextHandler(io.Discard, nil)))
	router := chi.NewRouter()
	router.Use(tokens.Middleware(errorsMiddleware, accounts, revoked))
	ProtectedRoutes(router, NewHandler(service), errorsMiddleware)
	router.Route("/users", func(r chi.Router) { users.Routes(r, users.NewHandler(users.NewService(accounts)), errorsMiddleware) })
	raw, err := tokens.Issue(userID, "ADMIN")
	require.NoError(t, err)
	other, err := tokens.Issue(userID, "ADMIN")
	require.NoError(t, err)
	request := func(method, path, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequestWithContext(t.Context(), method, path, nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	require.Equal(t, 200, request("GET", "/users/"+userID.String(), raw).Code)
	accounts.byEmail["user@example.com"].Role = users.RoleUser
	require.Equal(t, 403, request("GET", "/users/"+userID.String(), raw).Code)
	require.Equal(t, 200, request("GET", "/me", raw).Code)
	require.Equal(t, 204, request("POST", "/logout", raw).Code)
	require.Equal(t, 401, request("GET", "/me", raw).Code)
	require.Equal(t, 200, request("GET", "/me", other).Code)
	revoked.failure = errors.New("storage unavailable")
	w := request("GET", "/me", other)
	require.Equal(t, 500, w.Code)
	require.NotContains(t, w.Body.String(), "storage unavailable")
	revoked.failure = nil
	delete(accounts.byEmail, "user@example.com")
	require.Equal(t, 401, request("GET", "/me", other).Code)
}

func TestTokenRequiresExpiryIdentityAndSignature(t *testing.T) {
	m := NewTokenManager("signing-key", time.Hour)
	now := time.Now()
	for _, tc := range []struct {
		name   string
		claims jwt.RegisteredClaims
		method jwt.SigningMethod
	}{
		{"missing expiry", jwt.RegisteredClaims{Subject: uuid.New().String(), ID: uuid.New().String()}, jwt.SigningMethodHS256},
		{"missing jti", jwt.RegisteredClaims{Subject: uuid.New().String(), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}, jwt.SigningMethodHS256},
		{"expired", jwt.RegisteredClaims{Subject: uuid.New().String(), ID: uuid.New().String(), ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour))}, jwt.SigningMethodHS256},
		{"wrong algorithm", jwt.RegisteredClaims{Subject: uuid.New().String(), ID: uuid.New().String(), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}, jwt.SigningMethodHS384},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := jwt.NewWithClaims(tc.method, Claims{RegisteredClaims: tc.claims}).SignedString([]byte("signing-key"))
			require.NoError(t, err)
			_, err = m.Parse(raw)
			require.Error(t, err)
		})
	}
	raw, err := m.Issue(uuid.New(), "USER")
	require.NoError(t, err)
	_, err = NewTokenManager("different-key", time.Hour).Parse(raw)
	require.Error(t, err)
}

func TestRegistrationAndLoginPasswordBoundaries(t *testing.T) {
	accounts := newMemoryUsers()
	s := NewService(accounts, NewTokenManager("key", time.Hour), &memoryRevocations{entries: make(map[uuid.UUID]time.Time)})
	password := strings.Repeat("я", 36)
	session, err := s.Register(t.Context(), RegisterInput{Name: "User", Email: " User@Example.com ", Password: password})
	require.NoError(t, err)
	require.Equal(t, users.RoleUser, session.User.Role)
	require.NotEqual(t, password, session.User.PasswordHash)
	_, err = s.Login(t.Context(), LoginInput{Email: " USER@EXAMPLE.COM ", Password: password})
	require.NoError(t, err)
	_, err = s.Register(t.Context(), RegisterInput{Name: "Other", Email: "user@example.com", Password: password})
	require.ErrorIs(t, err, users.ErrEmailAlreadyExists)
	for _, pw := range []string{"short", strings.Repeat("a", 73), strings.Repeat("я", 37)} {
		_, err = s.Register(t.Context(), RegisterInput{Name: "User", Email: "new@example.com", Password: pw})
		require.ErrorIs(t, err, ErrInvalidPassword)
		_, err = s.Login(t.Context(), LoginInput{Email: "user@example.com", Password: pw})
		require.ErrorIs(t, err, ErrInvalidCredentials)
	}
	for _, email := range []string{"user@example.com", "missing@example.com"} {
		_, err = s.Login(t.Context(), LoginInput{Email: email, Password: "wrong-password"})
		require.ErrorIs(t, err, ErrInvalidCredentials)
	}
}
