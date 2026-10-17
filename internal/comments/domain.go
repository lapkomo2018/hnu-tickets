package comments

import (
	"time"
	"uuid"
)

type Comment struct {
	ID         uuid.UUID
	TicketID   uuid.UUID
	AuthorID   uuid.UUID
	AuthorName string
	Content    string
	CreatedAt  time.Time
}
