package auth

import (
	"net/http"

	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/users"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	var request registerRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	session, err := h.service.Register(r.Context(), RegisterInput(request))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toSessionResponse(session))
	return nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var request loginRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	session, err := h.service.Login(r.Context(), LoginInput(request))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toSessionResponse(session))
	return nil
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) error {
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	user, err := h.service.Me(r.Context(), actor.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, users.ToResponse(*user))
	return nil
}
