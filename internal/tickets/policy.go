package tickets

import (
	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/users"
)

func canView(actor appmiddleware.Principal, ticket *Ticket) bool {
	switch users.Role(actor.Role) {
	case users.RoleAdmin:
		return true
	case users.RoleUser:
		return ticket.CreatedBy == actor.UserID
	case users.RoleSpecialist:
		return ticket.AssignedTo != nil && *ticket.AssignedTo == actor.UserID
	default:
		return false
	}
}

func canEdit(actor appmiddleware.Principal, ticket *Ticket) bool {
	return users.Role(actor.Role) == users.RoleAdmin ||
		(users.Role(actor.Role) == users.RoleUser && ticket.CreatedBy == actor.UserID && ticket.Status == StatusNew)
}

func authorizeStatusChange(actor appmiddleware.Principal, ticket *Ticket, next Status) error {
	role := users.Role(actor.Role)
	if role == users.RoleAdmin {
		return nil
	}
	if role == users.RoleSpecialist && ticket.AssignedTo != nil && *ticket.AssignedTo == actor.UserID {
		if next == StatusInProgress || next == StatusWaitingForUser || next == StatusResolved {
			return nil
		}
	}
	if role == users.RoleUser && ticket.CreatedBy == actor.UserID {
		if next == StatusClosed || next == StatusCancelled || (ticket.Status == StatusWaitingForUser && next == StatusInProgress) {
			return nil
		}
	}
	return ErrStatusChangeForbidden
}

func forbiddenView() error {
	return ErrTicketAccessForbidden
}
