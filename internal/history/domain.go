package history

import (
	"time"
	"uuid"
)

const (
	ActionCreated         = "CREATED"
	ActionUpdated         = "UPDATED"
	ActionStatusChanged   = "STATUS_CHANGED"
	ActionAssigned        = "ASSIGNED"
	ActionReassigned      = "REASSIGNED"
	ActionPriorityChanged = "PRIORITY_CHANGED"
	ActionCategoryChanged = "CATEGORY_CHANGED"
	ActionCancelled       = "CANCELLED"
	ActionResolved        = "RESOLVED"
	ActionClosed          = "CLOSED"
)

type Entry struct {
	ID        uuid.UUID
	TicketID  uuid.UUID
	ChangedBy uuid.UUID
	Action    string
	OldValue  *string
	NewValue  *string
	CreatedAt time.Time
}
