package categories

import (
	"errors"

	"hnu-tickets/internal/platform/httpx"
)

const (
	codeCategoryNotFound   httpx.Code = "CATEGORY_NOT_FOUND"
	codeCategoryNameExists httpx.Code = "CATEGORY_NAME_EXISTS"
	codeCategoryInUse      httpx.Code = "CATEGORY_IN_USE"
	codeInvalidName        httpx.Code = "INVALID_CATEGORY_NAME"
	codeForbidden          httpx.Code = "FORBIDDEN"
)

func mapHTTPError(err error) error {
	switch {
	case errors.Is(err, ErrCategoryNotFound):
		return httpx.NotFound(codeCategoryNotFound, ErrCategoryNotFound.Error())
	case errors.Is(err, ErrCategoryNameExists):
		return httpx.Conflict(codeCategoryNameExists, ErrCategoryNameExists.Error())
	case errors.Is(err, ErrCategoryInUse):
		return httpx.Conflict(codeCategoryInUse, ErrCategoryInUse.Error())
	case errors.Is(err, ErrInvalidName):
		return httpx.BadRequest(codeInvalidName, ErrInvalidName.Error())
	case errors.Is(err, ErrAdminRequired):
		return httpx.Forbidden(codeForbidden, ErrAdminRequired.Error())
	default:
		return err
	}
}
