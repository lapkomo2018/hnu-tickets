package users

import (
	"errors"

	"hnu-tickets/internal/platform/httpx"
)

const (
	codeUserNotFound       httpx.Code = "USER_NOT_FOUND"
	codeEmailAlreadyExists httpx.Code = "EMAIL_ALREADY_EXISTS"
	codeInvalidName        httpx.Code = "INVALID_NAME"
	codeInvalidEmail       httpx.Code = "INVALID_EMAIL"
	codeInvalidRole        httpx.Code = "INVALID_ROLE"
	codeForbidden          httpx.Code = "FORBIDDEN"
)

func MapHTTPError(err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return httpx.NotFound(codeUserNotFound, ErrUserNotFound.Error())
	case errors.Is(err, ErrEmailAlreadyExists):
		return httpx.Conflict(codeEmailAlreadyExists, ErrEmailAlreadyExists.Error())
	case errors.Is(err, ErrInvalidName):
		return httpx.BadRequest(codeInvalidName, ErrInvalidName.Error())
	case errors.Is(err, ErrInvalidEmail):
		return httpx.BadRequest(codeInvalidEmail, ErrInvalidEmail.Error())
	case errors.Is(err, ErrInvalidRole):
		return httpx.BadRequest(codeInvalidRole, ErrInvalidRole.Error())
	case errors.Is(err, ErrAdminRequired):
		return httpx.Forbidden(codeForbidden, ErrAdminRequired.Error())
	default:
		return err
	}
}
