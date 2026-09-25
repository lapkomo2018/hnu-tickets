package users

import (
	"time"

	"github.com/google/uuid"
)

type Response struct {
	ID        uuid.UUID `json:"id" format:"uuid"`
	Name      string    `json:"name"`
	Email     string    `json:"email" format:"email"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"createdAt" format:"date-time"`
	UpdatedAt time.Time `json:"updatedAt" format:"date-time"`
}

type UpdateRequest struct {
	Name  *string `json:"name" minLength:"2" maxLength:"100"`
	Email *string `json:"email" format:"email"`
}

type UpdateRoleRequest struct {
	Role Role `json:"role" validate:"required"`
}

func ToResponse(user User) Response {
	return Response{
		ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role,
		CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
}
