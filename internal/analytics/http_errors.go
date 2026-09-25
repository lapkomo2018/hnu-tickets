package analytics

import (
	"errors"

	"hnu-tickets/internal/platform/httpx"
)

const (
	codeForbidden     httpx.Code = "FORBIDDEN"
	codeInvalidPeriod httpx.Code = "INVALID_PERIOD"
)

func mapHTTPError(err error) error {
	switch {
	case errors.Is(err, ErrAdminRequired):
		return httpx.Forbidden(codeForbidden, ErrAdminRequired.Error())
	case errors.Is(err, ErrInvalidPeriod):
		return httpx.BadRequest(codeInvalidPeriod, ErrInvalidPeriod.Error())
	default:
		return err
	}
}
