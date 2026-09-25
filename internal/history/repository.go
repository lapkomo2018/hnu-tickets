package history

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(context.Context, *Entry) error
	ListByTicket(context.Context, uuid.UUID) ([]Entry, error)
}
