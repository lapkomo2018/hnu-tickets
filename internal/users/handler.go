package users

import (
	"net/http"

	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	result, err := h.service.List(r.Context(), actor)
	if err != nil {
		return err
	}
	response := make([]Response, 0, len(result))
	for _, user := range result {
		response = append(response, ToResponse(user))
	}
	httpx.JSON(w, http.StatusOK, response)
	return nil
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	user, err := h.service.Get(r.Context(), actor, id)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, ToResponse(*user))
	return nil
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	var request UpdateRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	user, err := h.service.Update(r.Context(), actor, id, UpdateInput{Name: request.Name, Email: request.Email})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, ToResponse(*user))
	return nil
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	var request UpdateRoleRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	if err := h.service.UpdateRole(r.Context(), actor, id, request.Role); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
