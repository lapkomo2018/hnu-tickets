package auth

import (
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const PasswordCost = 12

func validPassword(password string) bool {
	return utf8.ValidString(password) && utf8.RuneCountInString(password) >= 8 && len(password) <= 72
}

func hashPassword(password string) (string, error) {
	if !validPassword(password) {
		return "", ErrInvalidPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	return string(hash), err
}

func checkPassword(hash, password string) error {
	if !validPassword(password) {
		return ErrInvalidCredentials
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
