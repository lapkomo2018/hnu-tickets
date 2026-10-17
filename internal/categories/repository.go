package categories

import (
	"context"
	"uuid"
)

type Repository interface {
	Create(context.Context, *Category) error
	GetByID(context.Context, uuid.UUID) (*Category, error)
	List(context.Context) ([]Category, error)
	Update(context.Context, *Category) error
	Delete(context.Context, uuid.UUID) error
}
