package auth

import (
	"errors"

	"hnu-tickets/internal/platform/httpx"
	"hnu-tickets/internal/users"
)

const (
	codeInvalidPassword    httpx.Code = "INVALID_PASSWORD"
	codeInvalidCredentials httpx.Code = "INVALID_CREDENTIALS"
)

func mapHTTPError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidPassword):
		return httpx.BadRequest(codeInvalidPassword, ErrInvalidPassword.Error())
	case errors.Is(err, ErrInvalidCredentials):
		return httpx.Unauthorized(codeInvalidCredentials, ErrInvalidCredentials.Error())
	default:
		return users.MapHTTPError(err)
	}
}
