package comments

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(context.Context, *Comment) error
	ListByTicket(context.Context, uuid.UUID) ([]Comment, error)
}
