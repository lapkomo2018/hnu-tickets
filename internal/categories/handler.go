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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	categories, err := h.service.List(r.Context())
	if err != nil {
		return err
	}
	response := make([]Response, 0, len(categories))
	for _, category := range categories {
		response = append(response, toResponse(category))
	}
	httpx.JSON(w, http.StatusOK, response)
	return nil
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	var requestBody Request
	if err := httpx.DecodeJSON(w, r, &requestBody); err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	category, err := h.service.Create(r.Context(), actor, Input(requestBody))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toResponse(*category))
	return nil
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	var requestBody Request
	if err := httpx.DecodeJSON(w, r, &requestBody); err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	category, err := h.service.Update(r.Context(), actor, id, Input(requestBody))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toResponse(*category))
	return nil
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	if err := h.service.Delete(r.Context(), actor, id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
