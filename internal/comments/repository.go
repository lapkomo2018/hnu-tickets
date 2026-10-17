package comments

import (
	"context"
	"uuid"
)

type Repository interface {
	Create(context.Context, *Comment) error
	ListByTicket(context.Context, uuid.UUID) ([]Comment, error)
}
