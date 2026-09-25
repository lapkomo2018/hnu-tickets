package history

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
)

type Handler struct {
	service *Service
}

type Response struct {
	ID        uuid.UUID `json:"id" format:"uuid"`
	TicketID  uuid.UUID `json:"ticketId" format:"uuid"`
	ChangedBy uuid.UUID `json:"changedBy" format:"uuid"`
	Action    string    `json:"action"`
	OldValue  *string   `json:"oldValue"`
	NewValue  *string   `json:"newValue"`
	CreatedAt time.Time `json:"createdAt" format:"date-time"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// @Summary List immutable ticket history
// @Tags History
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ticket identifier" format(uuid)
// @Success 200 {array} history.Response
// @Failure 400,401,403,404 {object} httpx.ErrorEnvelope
// @Router /tickets/{id}/history [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	ticketID, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	entries, err := h.service.List(r.Context(), actor, ticketID)
	if err != nil {
		return err
	}
	result := make([]Response, 0, len(entries))
	for _, entry := range entries {
		result = append(result, Response{
			ID: entry.ID, TicketID: entry.TicketID, ChangedBy: entry.ChangedBy,
			Action: entry.Action, OldValue: entry.OldValue, NewValue: entry.NewValue, CreatedAt: entry.CreatedAt,
		})
	}
	httpx.JSON(w, http.StatusOK, result)
	return nil
}
