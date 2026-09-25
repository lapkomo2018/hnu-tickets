package auth

import "errors"

var (
	ErrInvalidPassword    = errors.New("password must contain at least 8 characters")
	ErrInvalidCredentials = errors.New("email or password is incorrect")
)
