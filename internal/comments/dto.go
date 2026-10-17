package comments

import (
	"time"
	"uuid"
)

type CreateRequest struct {
	Content string `json:"content" validate:"required" minLength:"1" maxLength:"5000" example:"Please reconnect the printer"`
}

type Response struct {
	ID        uuid.UUID `json:"id" format:"uuid"`
	TicketID  uuid.UUID `json:"ticketId" format:"uuid"`
	Author    Author    `json:"Author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt" format:"date-time"`
}

type Author struct {
	ID   uuid.UUID `json:"id" format:"uuid"`
	Name string    `json:"name"`
}

func toResponse(comment Comment) Response {
	return Response{
		ID: comment.ID, TicketID: comment.TicketID,
		Author:  Author{ID: comment.AuthorID, Name: comment.AuthorName},
		Content: comment.Content, CreatedAt: comment.CreatedAt,
	}
}
