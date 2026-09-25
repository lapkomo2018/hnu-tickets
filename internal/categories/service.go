package categories

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/users"
)

type Service struct {
	repository Repository
}

type Input struct {
	Name        string
	Description string
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]Category, error) {
	return s.repository.List(ctx)
}

func (s *Service) Create(ctx context.Context, actor appmiddleware.Principal, input Input) (*Category, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	input, err := validateInput(input)
	if err != nil {
		return nil, err
	}
	category := &Category{Name: input.Name, Description: input.Description}
	if err := s.repository.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *Service) Update(ctx context.Context, actor appmiddleware.Principal, id uuid.UUID, input Input) (*Category, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	input, err := validateInput(input)
	if err != nil {
		return nil, err
	}
	category := &Category{ID: id, Name: input.Name, Description: input.Description}
	if err := s.repository.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *Service) Delete(ctx context.Context, actor appmiddleware.Principal, id uuid.UUID) error {
	if err := requireAdmin(actor); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

func validateInput(input Input) (Input, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if length := utf8.RuneCountInString(input.Name); length < 2 || length > 100 {
		return Input{}, ErrInvalidName
	}
	return input, nil
}

func requireAdmin(actor appmiddleware.Principal) error {
	if actor.Role != string(users.RoleAdmin) {
		return ErrAdminRequired
	}
	return nil
}
