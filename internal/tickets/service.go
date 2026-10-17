package tickets

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"hnu-tickets/internal/history"
	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/users"
)

type Service struct {
	repository Repository
	history    history.Repository
	users      users.Repository
	transactor Transactor
	now        func() time.Time
}

type CreateInput struct {
	Title       string
	Description string
	CategoryID  uuid.UUID
	Priority    Priority
}

type UpdateInput struct {
	Title       *string
	Description *string
	CategoryID  *uuid.UUID
	Priority    *Priority
}

func NewService(repository Repository, historyRepository history.Repository, usersRepository users.Repository, transactor Transactor) *Service {
	return &Service{
		repository: repository, history: historyRepository, users: usersRepository,
		transactor: transactor, now: time.Now,
	}
}

func (s *Service) Create(ctx context.Context, actor appmiddleware.Principal, input CreateInput) (*Record, error) {
	if users.Role(actor.Role) != users.RoleUser {
		return nil, ErrOnlyUsersCanCreate
	}
	input, err := validateCreate(input)
	if err != nil {
		return nil, err
	}
	ticket := &Ticket{
		Title: input.Title, Description: input.Description, CategoryID: input.CategoryID,
		Priority: input.Priority, Status: StatusNew, CreatedBy: actor.UserID,
	}
	err = s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repository.Create(txCtx, ticket); err != nil {
			return err
		}
		return s.history.Create(txCtx, &history.Entry{
			TicketID: ticket.ID, ChangedBy: actor.UserID, Action: history.ActionCreated,
		})
	})
	if err != nil {
		return nil, err
	}
	return s.repository.GetRecord(ctx, ticket.ID)
}

func (s *Service) Get(ctx context.Context, actor appmiddleware.Principal, id uuid.UUID) (*Record, error) {
	record, err := s.repository.GetRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canView(actor, &record.Ticket) {
		return nil, forbiddenView()
	}
	return record, nil
}

func (s *Service) List(ctx context.Context, actor appmiddleware.Principal, filter Filter) (Page, error) {
	if err := normalizeFilter(&filter); err != nil {
		return Page{}, err
	}
	switch users.Role(actor.Role) {
	case users.RoleUser:
		filter.CreatedBy = &actor.UserID
	case users.RoleSpecialist:
		filter.AssignedTo = &actor.UserID
	case users.RoleAdmin:
	default:
		return Page{}, forbiddenView()
	}
	items, total, err := s.repository.List(ctx, filter)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, Page: filter.Page, PageSize: filter.PageSize, Total: total}, nil
}

func (s *Service) Update(ctx context.Context, actor appmiddleware.Principal, id uuid.UUID, input UpdateInput) (*Record, error) {
	err := s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		ticket, err := s.repository.GetByIDForUpdate(txCtx, id)
		if err != nil {
			return err
		}
		if !canEdit(actor, ticket) {
			return ErrTicketCannotBeEdited
		}
		entries, err := applyUpdate(ticket, actor.UserID, input, s.now().UTC())
		if err != nil {
			return err
		}
		if err := s.repository.Update(txCtx, ticket); err != nil {
			return err
		}
		for i := range entries {
			if err := s.history.Create(txCtx, &entries[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.repository.GetRecord(ctx, id)
}

func (s *Service) Assign(ctx context.Context, actor appmiddleware.Principal, id, specialistID uuid.UUID) (*Record, error) {
	if users.Role(actor.Role) != users.RoleAdmin {
		return nil, ErrOnlyAdminsCanAssign
	}
	err := s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		specialist, err := s.users.GetByID(txCtx, specialistID)
		if err != nil {
			if errors.Is(err, users.ErrUserNotFound) {
				return ErrInvalidSpecialist
			}
			return err
		}
		if specialist.Role != users.RoleSpecialist {
			return ErrSpecialistRequired
		}
		ticket, err := s.repository.GetByIDForUpdate(txCtx, id)
		if err != nil {
			return err
		}
		if ticket.Status.Terminal() {
			return ErrTerminalTicket
		}
		var oldValue *string
		action := history.ActionAssigned
		if ticket.AssignedTo != nil {
			value := ticket.AssignedTo.String()
			oldValue = &value
			action = history.ActionReassigned
		}
		ticket.AssignedTo = &specialistID
		ticket.UpdatedAt = s.now().UTC()
		if ticket.Status == StatusNew {
			if err := ticket.ChangeStatus(StatusAssigned, s.now()); err != nil {
				return err
			}
		}
		if err := s.repository.Update(txCtx, ticket); err != nil {
			return err
		}
		newValue := specialistID.String()
		return s.history.Create(txCtx, &history.Entry{
			TicketID: id, ChangedBy: actor.UserID, Action: action, OldValue: oldValue, NewValue: &newValue,
		})
	})
	if err != nil {
		return nil, err
	}
	return s.repository.GetRecord(ctx, id)
}

func (s *Service) ChangeStatus(ctx context.Context, actor appmiddleware.Principal, id uuid.UUID, next Status) (*Record, error) {
	if !next.Valid() {
		return nil, ErrInvalidStatus
	}
	err := s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		ticket, err := s.repository.GetByIDForUpdate(txCtx, id)
		if err != nil {
			return err
		}
		if err := authorizeStatusChange(actor, ticket, next); err != nil {
			return err
		}
		oldStatus := ticket.Status
		if err := ticket.ChangeStatus(next, s.now()); err != nil {
			return err
		}
		if err := s.repository.Update(txCtx, ticket); err != nil {
			return err
		}
		oldValue, newValue := string(oldStatus), string(next)
		return s.history.Create(txCtx, &history.Entry{
			TicketID: id, ChangedBy: actor.UserID, Action: statusAction(next),
			OldValue: &oldValue, NewValue: &newValue,
		})
	})
	if err != nil {
		return nil, err
	}
	return s.repository.GetRecord(ctx, id)
}

func (s *Service) AuthorizeView(ctx context.Context, actor appmiddleware.Principal, id uuid.UUID) error {
	ticket, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !canView(actor, ticket) {
		return forbiddenView()
	}
	return nil
}

func validateCreate(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if length := utf8.RuneCountInString(input.Title); length < 3 || length > 255 {
		return CreateInput{}, ErrInvalidTitle
	}
	if input.Description == "" {
		return CreateInput{}, ErrInvalidDescription
	}
	if input.CategoryID == uuid.Nil() {
		return CreateInput{}, ErrCategoryRequired
	}
	if !input.Priority.Valid() {
		return CreateInput{}, ErrInvalidPriority
	}
	return input, nil
}

func normalizeFilter(filter *Filter) error {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	if filter.Status != nil && !filter.Status.Valid() {
		return ErrInvalidStatus
	}
	if filter.Priority != nil && !filter.Priority.Valid() {
		return ErrInvalidPriority
	}
	if filter.Order != "" && !strings.EqualFold(filter.Order, "asc") && !strings.EqualFold(filter.Order, "desc") {
		return ErrInvalidOrder
	}
	if filter.Sort != "" {
		allowed := map[string]bool{"createdAt": true, "updatedAt": true, "priority": true, "status": true}
		if !allowed[filter.Sort] {
			return ErrInvalidSort
		}
	}
	return nil
}

func applyUpdate(ticket *Ticket, actorID uuid.UUID, input UpdateInput, now time.Time) ([]history.Entry, error) {
	entries := make([]history.Entry, 0, 4)
	add := func(action, oldValue, newValue string) {
		oldCopy, newCopy := oldValue, newValue
		entries = append(entries, history.Entry{
			TicketID: ticket.ID, ChangedBy: actorID, Action: action, OldValue: &oldCopy, NewValue: &newCopy,
		})
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if length := utf8.RuneCountInString(title); length < 3 || length > 255 {
			return nil, ErrInvalidTitle
		}
		if title != ticket.Title {
			add(history.ActionUpdated, ticket.Title, title)
			ticket.Title = title
		}
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		if description == "" {
			return nil, ErrInvalidDescription
		}
		if description != ticket.Description {
			add(history.ActionUpdated, ticket.Description, description)
			ticket.Description = description
		}
	}
	if input.Priority != nil {
		if !input.Priority.Valid() {
			return nil, ErrInvalidPriority
		}
		if *input.Priority != ticket.Priority {
			add(history.ActionPriorityChanged, string(ticket.Priority), string(*input.Priority))
			ticket.Priority = *input.Priority
		}
	}
	if input.CategoryID != nil && *input.CategoryID != ticket.CategoryID {
		add(history.ActionCategoryChanged, ticket.CategoryID.String(), input.CategoryID.String())
		ticket.CategoryID = *input.CategoryID
	}
	ticket.UpdatedAt = now
	return entries, nil
}

func statusAction(status Status) string {
	switch status {
	case StatusCancelled:
		return history.ActionCancelled
	case StatusResolved:
		return history.ActionResolved
	case StatusClosed:
		return history.ActionClosed
	default:
		return history.ActionStatusChanged
	}
}
