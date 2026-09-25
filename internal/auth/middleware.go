package auth

import (
	"net/http"
	"strings"

	"hnu-tickets/internal/platform/apperror"
	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
)

func (m *TokenManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			httpx.Error(w, r, apperror.New(apperror.KindUnauthorized, "UNAUTHORIZED", "a bearer token is required"))
			return
		}
		principal, err := m.Parse(parts[1])
		if err != nil {
			httpx.Error(w, r, apperror.New(apperror.KindUnauthorized, "INVALID_TOKEN", "access token is invalid or expired"))
			return
		}
		next.ServeHTTP(w, appmiddleware.WithPrincipal(r, principal))
	})
}
