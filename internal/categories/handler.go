package categories

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
	categories, err := h.service.List(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	response := make([]Response, 0, len(categories))
	for _, category := range categories {
		response = append(response, toResponse(category))
	}
	httpx.JSON(w, http.StatusOK, response)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var requestBody request
	if err := httpx.DecodeJSON(w, r, &requestBody); err != nil {
		httpx.Error(w, r, err)
		return
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	category, err := h.service.Create(r.Context(), actor, Input(requestBody))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toResponse(*category))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	var requestBody request
	if err := httpx.DecodeJSON(w, r, &requestBody); err != nil {
		httpx.Error(w, r, err)
		return
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	category, err := h.service.Update(r.Context(), actor, id, Input(requestBody))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toResponse(*category))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
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
	if err := h.service.Delete(r.Context(), actor, id); err != nil {
		httpx.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
