package comments

import (
	"context"
	"strings"
	"unicode/utf8"
	"uuid"

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

func (s *Service) List(ctx context.Context, actor appmiddleware.Principal, ticketID uuid.UUID) ([]Comment, error) {
	if err := s.access.AuthorizeView(ctx, actor, ticketID); err != nil {
		return nil, err
	}
	return s.repository.ListByTicket(ctx, ticketID)
}

func (s *Service) Create(ctx context.Context, actor appmiddleware.Principal, ticketID uuid.UUID, content string) (*Comment, error) {
	if err := s.access.AuthorizeView(ctx, actor, ticketID); err != nil {
		return nil, err
	}
	content = strings.TrimSpace(content)
	if length := utf8.RuneCountInString(content); length < 1 || length > 5000 {
		return nil, ErrInvalidComment
	}
	comment := &Comment{TicketID: ticketID, AuthorID: actor.UserID, Content: content}
	if err := s.repository.Create(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}
