package httpx

import (
	"net/http"
	"uuid"

	"github.com/go-chi/chi/v5"
)

func UUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	value, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil(), NewError(http.StatusBadRequest, CodeInvalidID, name+" must be a UUID")
	}
	return value, nil
}
