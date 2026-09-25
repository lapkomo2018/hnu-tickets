package tickets

import (
	"github.com/go-chi/chi/v5"

	"hnu-tickets/internal/comments"
	"hnu-tickets/internal/history"
	"hnu-tickets/internal/platform/httpx"
)

func Routes(
	router chi.Router,
	handler *Handler,
	commentsHandler *comments.Handler,
	historyHandler *history.Handler,
	errors *httpx.ErrorMiddleware,
) {
	handle := errors.HandleWith(mapHTTPError)
	router.Get("/", handle(handler.List))
	router.Post("/", handle(handler.Create))
	router.Get("/{id}", handle(handler.Get))
	router.Patch("/{id}", handle(handler.Update))
	router.Patch("/{id}/assign", handle(handler.Assign))
	router.Patch("/{id}/status", handle(handler.ChangeStatus))
	router.Get("/{id}/comments", handle(commentsHandler.List))
	router.Post("/{id}/comments", handle(commentsHandler.Create))
	router.Get("/{id}/history", handle(historyHandler.List))
}
