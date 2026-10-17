package users

import (
	"context"
	"net/mail"
	"strings"
	"unicode/utf8"
	"uuid"

	appmiddleware "hnu-tickets/internal/platform/middleware"
)

type Service struct {
	repository Repository
}

type UpdateInput struct {
	Name  *string
	Email *string
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, actor appmiddleware.Principal) ([]User, error) {
	if actor.Role != string(RoleAdmin) {
		return nil, forbidden()
	}
	return s.repository.List(ctx)
}

func (s *Service) Get(ctx context.Context, actor appmiddleware.Principal, id uuid.UUID) (*User, error) {
	if actor.Role != string(RoleAdmin) {
		return nil, forbidden()
	}
	return s.repository.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, actor appmiddleware.Principal, id uuid.UUID, input UpdateInput) (*User, error) {
	if actor.Role != string(RoleAdmin) {
		return nil, forbidden()
	}
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if length := utf8.RuneCountInString(name); length < 2 || length > 100 {
			return nil, ErrInvalidName
		}
		user.Name = name
	}
	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		address, err := mail.ParseAddress(email)
		if err != nil || address.Address != email {
			return nil, ErrInvalidEmail
		}
		user.Email = email
	}
	if err := s.repository.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateRole(ctx context.Context, actor appmiddleware.Principal, id uuid.UUID, role Role) error {
	if actor.Role != string(RoleAdmin) {
		return forbidden()
	}
	if !role.Valid() {
		return ErrInvalidRole
	}
	return s.repository.UpdateRole(ctx, id, role)
}

func forbidden() error {
	return ErrAdminRequired
}
