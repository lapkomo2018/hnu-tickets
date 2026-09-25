package comments

import (
	"time"

	"github.com/google/uuid"
)

type createRequest struct {
	Content string `json:"content"`
}

type response struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticketId"`
	Author    author    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type author struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func toResponse(comment Comment) response {
	return response{
		ID: comment.ID, TicketID: comment.TicketID,
		Author:  author{ID: comment.AuthorID, Name: comment.AuthorName},
		Content: comment.Content, CreatedAt: comment.CreatedAt,
	}
}
