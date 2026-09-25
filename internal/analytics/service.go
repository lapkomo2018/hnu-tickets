package analytics

import (
	"context"

	"hnu-tickets/internal/platform/apperror"
	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/users"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Overview(ctx context.Context, actor appmiddleware.Principal, period Period) (Overview, error) {
	if err := authorize(actor, period); err != nil {
		return Overview{}, err
	}
	return s.repository.Overview(ctx, period)
}

func (s *Service) ByStatus(ctx context.Context, actor appmiddleware.Principal, period Period) ([]Count, error) {
	if err := authorize(actor, period); err != nil {
		return nil, err
	}
	return s.repository.ByStatus(ctx, period)
}

func (s *Service) ByPriority(ctx context.Context, actor appmiddleware.Principal, period Period) ([]Count, error) {
	if err := authorize(actor, period); err != nil {
		return nil, err
	}
	return s.repository.ByPriority(ctx, period)
}

func (s *Service) ByCategory(ctx context.Context, actor appmiddleware.Principal, period Period) ([]CategoryCount, error) {
	if err := authorize(actor, period); err != nil {
		return nil, err
	}
	return s.repository.ByCategory(ctx, period)
}

func (s *Service) Specialists(ctx context.Context, actor appmiddleware.Principal, period Period) ([]SpecialistWorkload, error) {
	if err := authorize(actor, period); err != nil {
		return nil, err
	}
	return s.repository.Specialists(ctx, period)
}

func authorize(actor appmiddleware.Principal, period Period) error {
	if users.Role(actor.Role) != users.RoleAdmin {
		return apperror.New(apperror.KindForbidden, "FORBIDDEN", "administrator access is required")
	}
	if period.From != nil && period.To != nil && period.From.After(*period.To) {
		return apperror.New(apperror.KindValidation, "INVALID_PERIOD", "from must not be after to")
	}
	return nil
}
