package comments

import (
	"errors"

	"hnu-tickets/internal/platform/httpx"
)

const codeInvalidComment httpx.Code = "INVALID_COMMENT"

func MapHTTPError(err error) error {
	if errors.Is(err, ErrInvalidComment) {
		return httpx.BadRequest(codeInvalidComment, ErrInvalidComment.Error())
	}
	return err
}
