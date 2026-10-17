package middleware

import (
	"context"
	"net/http"
	"time"
	"uuid"

	"hnu-tickets/internal/platform/httpx"
)

type Principal struct {
	UserID    uuid.UUID
	Role      string
	TokenID   uuid.UUID
	ExpiresAt time.Time
}

type principalKey struct{}

func WithPrincipal(r *http.Request, principal Principal) *http.Request {
	ctx := context.WithValue(r.Context(), principalKey{}, principal)
	return r.WithContext(ctx)
}

func PrincipalFromContext(ctx context.Context) (Principal, error) {
	principal, ok := ctx.Value(principalKey{}).(Principal)
	if !ok {
		return Principal{}, httpx.Unauthorized(httpx.CodeUnauthorized, "authentication is required")
	}
	return principal, nil
}
