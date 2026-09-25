package users

import (
	"time"

	"github.com/google/uuid"
)

type Response struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

type UpdateRoleRequest struct {
	Role Role `json:"role"`
}

func ToResponse(user User) Response {
	return Response{
		ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role,
		CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
}
