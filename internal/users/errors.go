package users

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("a user with this email already exists")
	ErrInvalidName        = errors.New("name must contain 2 to 100 characters")
	ErrInvalidEmail       = errors.New("email address is invalid")
	ErrInvalidRole        = errors.New("role is invalid")
	ErrAdminRequired      = errors.New("administrator access is required")
)
