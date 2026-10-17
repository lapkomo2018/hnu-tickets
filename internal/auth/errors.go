package auth

import "errors"

var (
	ErrInvalidPassword    = errors.New("password must contain at least 8 characters and at most 72 UTF-8 bytes")
	ErrInvalidCredentials = errors.New("email or password is incorrect")
)
