package users

import (
	"context"
	"uuid"
)

type Repository interface {
	Create(context.Context, *User) error
	GetByID(context.Context, uuid.UUID) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
	List(context.Context) ([]User, error)
	Update(context.Context, *User) error
	UpdateRole(context.Context, uuid.UUID, Role) error
}
