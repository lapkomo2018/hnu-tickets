package tickets

import (
	"time"
	"uuid"
)

type Ticket struct {
	ID          uuid.UUID
	Title       string
	Description string
	Status      Status
	Priority    Priority
	CategoryID  uuid.UUID
	CreatedBy   uuid.UUID
	AssignedTo  *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ResolvedAt  *time.Time
	ClosedAt    *time.Time
}

func (t *Ticket) ChangeStatus(next Status, now time.Time) error {
	_, ok := transitions[t.Status][next]
	if !ok {
		return &InvalidStatusTransitionError{From: t.Status, To: next}
	}

	previous := t.Status
	t.Status = next
	t.UpdatedAt = now.UTC()
	if next == StatusResolved {
		resolvedAt := now.UTC()
		t.ResolvedAt = &resolvedAt
	}
	if previous == StatusResolved && next == StatusInProgress {
		t.ResolvedAt = nil
	}
	if next == StatusClosed {
		closedAt := now.UTC()
		t.ClosedAt = &closedAt
	}
	return nil
}
