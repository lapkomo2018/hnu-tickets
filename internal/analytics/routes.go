package analytics

import "github.com/go-chi/chi/v5"

func Routes(router chi.Router, handler *Handler) {
	router.Get("/overview", handler.Overview)
	router.Get("/statuses", handler.Statuses)
	router.Get("/priorities", handler.Priorities)
	router.Get("/categories", handler.Categories)
	router.Get("/specialists", handler.Specialists)
}
