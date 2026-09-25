package tickets

import (
	"github.com/go-chi/chi/v5"

	"hnu-tickets/internal/comments"
	"hnu-tickets/internal/history"
)

func Routes(router chi.Router, handler *Handler, commentsHandler *comments.Handler, historyHandler *history.Handler) {
	router.Get("/", handler.List)
	router.Post("/", handler.Create)
	router.Get("/{id}", handler.Get)
	router.Patch("/{id}", handler.Update)
	router.Patch("/{id}/assign", handler.Assign)
	router.Patch("/{id}/status", handler.ChangeStatus)
	router.Get("/{id}/comments", commentsHandler.List)
	router.Post("/{id}/comments", commentsHandler.Create)
	router.Get("/{id}/history", historyHandler.List)
}
