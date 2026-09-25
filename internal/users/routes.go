package users

import "github.com/go-chi/chi/v5"

func Routes(router chi.Router, handler *Handler) {
	router.Get("/", handler.List)
	router.Get("/{id}", handler.Get)
	router.Patch("/{id}", handler.Update)
	router.Patch("/{id}/role", handler.UpdateRole)
}
