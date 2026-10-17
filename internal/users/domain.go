package users

import (
	"time"
	"uuid"
)

type Role string

const (
	RoleUser       Role = "USER"
	RoleSpecialist Role = "SPECIALIST"
	RoleAdmin      Role = "ADMIN"
)

func (r Role) Valid() bool {
	return r == RoleUser || r == RoleSpecialist || r == RoleAdmin
}

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
