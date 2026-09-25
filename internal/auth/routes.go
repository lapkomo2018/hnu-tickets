package auth

import (
	"github.com/go-chi/chi/v5"

	"hnu-tickets/internal/platform/httpx"
)

func PublicRoutes(router chi.Router, handler *Handler, errors *httpx.ErrorMiddleware) {
	handle := errors.HandleWith(mapHTTPError)
	router.Post("/register", handle(handler.Register))
	router.Post("/login", handle(handler.Login))
}

func ProtectedRoutes(router chi.Router, handler *Handler, errors *httpx.ErrorMiddleware) {
	handle := errors.HandleWith(mapHTTPError)
	router.Get("/me", handle(handler.Me))
}
