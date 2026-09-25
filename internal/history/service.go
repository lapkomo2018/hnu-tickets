package history

import (
	"context"

	"github.com/google/uuid"

	appmiddleware "hnu-tickets/internal/platform/middleware"
)

type AccessChecker interface {
	AuthorizeView(context.Context, appmiddleware.Principal, uuid.UUID) error
}

type Service struct {
	repository Repository
	access     AccessChecker
}

func NewService(repository Repository, access AccessChecker) *Service {
	return &Service{repository: repository, access: access}
}

func (s *Service) List(ctx context.Context, actor appmiddleware.Principal, ticketID uuid.UUID) ([]Entry, error) {
	if err := s.access.AuthorizeView(ctx, actor, ticketID); err != nil {
		return nil, err
	}
	return s.repository.ListByTicket(ctx, ticketID)
}
