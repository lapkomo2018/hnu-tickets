package categories

import (
	"github.com/go-chi/chi/v5"

	"hnu-tickets/internal/platform/httpx"
)

func Routes(router chi.Router, handler *Handler, errors *httpx.ErrorMiddleware) {
	handle := errors.HandleWith(mapHTTPError)
	router.Get("/", handle(handler.List))
	router.Post("/", handle(handler.Create))
	router.Patch("/{id}", handle(handler.Update))
	router.Delete("/{id}", handle(handler.Delete))
}
