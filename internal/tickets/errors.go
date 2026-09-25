package tickets

import (
	"errors"
	"fmt"
)

var (
	ErrTicketNotFound        = errors.New("ticket not found")
	ErrInvalidTitle          = errors.New("title must contain 3 to 255 characters")
	ErrInvalidDescription    = errors.New("description is required")
	ErrCategoryRequired      = errors.New("categoryId is required")
	ErrInvalidCategory       = errors.New("category does not exist")
	ErrInvalidReference      = errors.New("category or assigned user does not exist")
	ErrInvalidPriority       = errors.New("priority is invalid")
	ErrInvalidStatus         = errors.New("status is invalid")
	ErrInvalidOrder          = errors.New("order must be asc or desc")
	ErrInvalidSort           = errors.New("sort field is invalid")
	ErrInvalidSpecialist     = errors.New("specialist does not exist")
	ErrSpecialistRequired    = errors.New("assigned user must be a specialist")
	ErrTerminalTicket        = errors.New("closed or cancelled tickets cannot be assigned")
	ErrOnlyUsersCanCreate    = errors.New("only users can create tickets")
	ErrTicketCannotBeEdited  = errors.New("this ticket cannot be edited")
	ErrOnlyAdminsCanAssign   = errors.New("only administrators can assign tickets")
	ErrStatusChangeForbidden = errors.New("you cannot change this ticket status")
	ErrTicketAccessForbidden = errors.New("you cannot access this ticket")
)

type InvalidStatusTransitionError struct {
	From Status
	To   Status
}

func (e *InvalidStatusTransitionError) Error() string {
	return fmt.Sprintf("cannot transition ticket from %s to %s", e.From, e.To)
}
