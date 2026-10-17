package auth

import (
	stderrors "errors"
	"net/http"
	"strings"

	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/users"
)

func (m *TokenManager) Middleware(errors *httpx.ErrorMiddleware, accounts users.Repository, revoked Revocations) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				errors.Write(w, r, httpx.NewError(http.StatusUnauthorized, httpx.CodeUnauthorized, "a bearer token is required"))
				return
			}
			principal, err := m.Parse(parts[1])
			if err != nil {
				errors.Write(w, r, httpx.NewError(http.StatusUnauthorized, httpx.CodeInvalidToken, "access token is invalid or expired"))
				return
			}
			blocked, err := revoked.IsRevoked(r.Context(), principal.TokenID)
			if err != nil {
				errors.Write(w, r, err)
				return
			}
			if blocked {
				errors.Write(w, r, httpx.Unauthorized(httpx.CodeInvalidToken, "access token is invalid or expired"))
				return
			}
			user, err := accounts.GetByID(r.Context(), principal.UserID)
			if stderrors.Is(err, users.ErrUserNotFound) || (err == nil && !user.Role.Valid()) {
				errors.Write(w, r, httpx.Unauthorized(httpx.CodeInvalidToken, "access token is invalid or expired"))
				return
			}
			if err != nil {
				errors.Write(w, r, err)
				return
			}
			principal.Role = string(user.Role)
			next.ServeHTTP(w, appmiddleware.WithPrincipal(r, principal))
		})
	}
}
