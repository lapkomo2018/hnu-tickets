package comments

import (
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	Content string `json:"content"`
}

type Response struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticketId"`
	Author    Author    `json:"Author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type Author struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func toResponse(comment Comment) Response {
	return Response{
		ID: comment.ID, TicketID: comment.TicketID,
		Author:  Author{ID: comment.AuthorID, Name: comment.AuthorName},
		Content: comment.Content, CreatedAt: comment.CreatedAt,
	}
}
