package users

import (
	"github.com/go-chi/chi/v5"

	"hnu-tickets/internal/platform/httpx"
)

func Routes(router chi.Router, handler *Handler, errors *httpx.ErrorMiddleware) {
	handle := errors.HandleWith(MapHTTPError)
	router.Get("/", handle(handler.List))
	router.Get("/{id}", handle(handler.Get))
	router.Patch("/{id}", handle(handler.Update))
	router.Patch("/{id}/role", handle(handler.UpdateRole))
}
