package categories

import "github.com/go-chi/chi/v5"

func Routes(router chi.Router, handler *Handler) {
	router.Get("/", handler.List)
	router.Post("/", handler.Create)
	router.Patch("/{id}", handler.Update)
	router.Delete("/{id}", handler.Delete)
}
