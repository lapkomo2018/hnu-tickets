package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"hnu-tickets/internal/platform/apperror"
)

type Principal struct {
	UserID uuid.UUID
	Role   string
}

type principalKey struct{}

func WithPrincipal(r *http.Request, principal Principal) *http.Request {
	ctx := context.WithValue(r.Context(), principalKey{}, principal)
	return r.WithContext(ctx)
}

func PrincipalFromContext(ctx context.Context) (Principal, error) {
	principal, ok := ctx.Value(principalKey{}).(Principal)
	if !ok {
		return Principal{}, apperror.New(apperror.KindUnauthorized, "UNAUTHORIZED", "authentication is required")
	}
	return principal, nil
}
