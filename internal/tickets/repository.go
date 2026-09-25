package tickets

import (
	"context"

	"github.com/google/uuid"
)

type UserSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CategorySummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Record struct {
	Ticket
	Category     CategorySummary
	Creator      UserSummary
	AssigneeName *string
}

type Repository interface {
	Create(context.Context, *Ticket) error
	GetByID(context.Context, uuid.UUID) (*Ticket, error)
	GetByIDForUpdate(context.Context, uuid.UUID) (*Ticket, error)
	GetRecord(context.Context, uuid.UUID) (*Record, error)
	List(context.Context, Filter) ([]Record, int64, error)
	Update(context.Context, *Ticket) error
}

type Transactor interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}
