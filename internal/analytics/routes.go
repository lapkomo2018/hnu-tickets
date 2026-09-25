package analytics

import (
	"github.com/go-chi/chi/v5"

	"hnu-tickets/internal/platform/httpx"
)

func Routes(router chi.Router, handler *Handler, errors *httpx.ErrorMiddleware) {
	handle := errors.HandleWith(mapHTTPError)
	router.Get("/overview", handle(handler.Overview))
	router.Get("/statuses", handle(handler.Statuses))
	router.Get("/priorities", handle(handler.Priorities))
	router.Get("/categories", handle(handler.Categories))
	router.Get("/specialists", handle(handler.Specialists))
}
