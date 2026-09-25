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

type response struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticketId"`
	ChangedBy uuid.UUID `json:"changedBy"`
	Action    string    `json:"action"`
	OldValue  *string   `json:"oldValue"`
	NewValue  *string   `json:"newValue"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

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
	result := make([]response, 0, len(entries))
	for _, entry := range entries {
		result = append(result, response{
			ID: entry.ID, TicketID: entry.TicketID, ChangedBy: entry.ChangedBy,
			Action: entry.Action, OldValue: entry.OldValue, NewValue: entry.NewValue, CreatedAt: entry.CreatedAt,
		})
	}
	httpx.JSON(w, http.StatusOK, result)
	return nil
}
