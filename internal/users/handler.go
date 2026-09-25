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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	result, err := h.service.List(r.Context(), actor)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	response := make([]Response, 0, len(result))
	for _, user := range result {
		response = append(response, ToResponse(user))
	}
	httpx.JSON(w, http.StatusOK, response)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	user, err := h.service.Get(r.Context(), actor, id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, ToResponse(*user))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	var request updateRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.Error(w, r, err)
		return
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	user, err := h.service.Update(r.Context(), actor, id, UpdateInput{Name: request.Name, Email: request.Email})
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, ToResponse(*user))
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	var request updateRoleRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.Error(w, r, err)
		return
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	if err := h.service.UpdateRole(r.Context(), actor, id, request.Role); err != nil {
		httpx.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
