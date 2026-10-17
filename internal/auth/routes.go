package auth

import (
	"time"

	"github.com/go-chi/chi/v5"

	"hnu-tickets/internal/platform/httpx"
)

func PublicRoutes(router chi.Router, handler *Handler, errors *httpx.ErrorMiddleware) {
	handle := errors.HandleWith(mapHTTPError)
	router.With(NewAttemptLimiter(5, time.Minute).Middleware(errors)).Post("/register", handle(handler.Register))
	router.With(NewAttemptLimiter(10, time.Minute).Middleware(errors)).Post("/login", handle(handler.Login))
}

func ProtectedRoutes(router chi.Router, handler *Handler, errors *httpx.ErrorMiddleware) {
	handle := errors.HandleWith(mapHTTPError)
	router.Get("/me", handle(handler.Me))
	router.Post("/logout", handle(handler.Logout))
}
