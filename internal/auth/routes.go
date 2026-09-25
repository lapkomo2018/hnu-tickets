package auth

import "github.com/go-chi/chi/v5"

func PublicRoutes(router chi.Router, handler *Handler) {
	router.Post("/register", handler.Register)
	router.Post("/login", handler.Login)
}

func ProtectedRoutes(router chi.Router, handler *Handler) {
	router.Get("/me", handler.Me)
}
