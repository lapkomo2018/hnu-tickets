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

// @Summary List ticket comments
// @Tags Comments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket identifier" format(uuid)
// @Success 200 {array} comments.Response
// @Failure 400,401,403,404 {object} httpx.ErrorEnvelope
// @Router /tickets/{id}/comments [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	ticketID, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	comments, err := h.service.List(r.Context(), actor, ticketID)
	if err != nil {
		return err
	}
	result := make([]Response, 0, len(comments))
	for _, comment := range comments {
		result = append(result, toResponse(comment))
	}
	httpx.JSON(w, http.StatusOK, result)
	return nil
}

// @Summary Add a ticket comment
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket identifier" format(uuid)
// @Param request body comments.CreateRequest true "Comment content"
// @Success 201 {object} comments.Response
// @Failure 400,401,403,404 {object} httpx.ErrorEnvelope
// @Router /tickets/{id}/comments [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	ticketID, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	var request CreateRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	comment, err := h.service.Create(r.Context(), actor, ticketID, request.Content)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toResponse(*comment))
	return nil
}
