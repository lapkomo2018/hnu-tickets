package comments

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
	ticketID, err := httpx.UUIDParam(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	comments, err := h.service.List(r.Context(), actor, ticketID)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	result := make([]response, 0, len(comments))
	for _, comment := range comments {
		result = append(result, toResponse(comment))
	}
	httpx.JSON(w, http.StatusOK, result)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ticketID, err := httpx.UUIDParam(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	var request createRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.Error(w, r, err)
		return
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	comment, err := h.service.Create(r.Context(), actor, ticketID, request.Content)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toResponse(*comment))
}
