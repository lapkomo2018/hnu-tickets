package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
	"uuid"

	"hnu-tickets/internal/users"
)

// A public dummy hash equalizes password work for unknown accounts.
const dummyPasswordHash = "$2b$12$P8Sa9EaN/QNferpWlJ/gI.eZZdN7c6sH7SgzgqcovOiODAJWpHvbq"

type Service struct {
	users       users.Repository
	tokens      *TokenManager
	revocations Revocations
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type Session struct {
	AccessToken string
	User        users.User
}

func NewService(repository users.Repository, tokens *TokenManager, revocations Revocations) *Service {
	return &Service{users: repository, tokens: tokens, revocations: revocations}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*Session, error) {
	name, email, err := validateCredentials(input.Name, input.Email, input.Password)
	if err != nil {
		return nil, err
	}
	hash, err := hashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user := &users.User{Name: name, Email: email, PasswordHash: hash, Role: users.RoleUser}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return s.sessionFor(user)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*Session, error) {
	if !validPassword(input.Password) {
		return nil, invalidCredentials()
	}
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			_ = checkPassword(dummyPasswordHash, input.Password)
			return nil, invalidCredentials()
		}
		return nil, err
	}
	if checkPassword(user.PasswordHash, input.Password) != nil {
		return nil, invalidCredentials()
	}
	return s.sessionFor(user)
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*users.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *Service) sessionFor(user *users.User) (*Session, error) {
	token, err := s.tokens.Issue(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}
	return &Session{AccessToken: token, User: *user}, nil
}

func validateCredentials(name, email, password string) (string, string, error) {
	name = strings.TrimSpace(name)
	if length := utf8.RuneCountInString(name); length < 2 || length > 100 {
		return "", "", users.ErrInvalidName
	}
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", "", users.ErrInvalidEmail
	}
	if !validPassword(password) {
		return "", "", ErrInvalidPassword
	}
	return name, email, nil
}

func invalidCredentials() error {
	return ErrInvalidCredentials
}
