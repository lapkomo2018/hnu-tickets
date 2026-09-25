package tickets

import (
	"errors"

	"hnu-tickets/internal/comments"
	"hnu-tickets/internal/platform/httpx"
)

const (
	codeTicketNotFound          httpx.Code = "TICKET_NOT_FOUND"
	codeInvalidTitle            httpx.Code = "INVALID_TITLE"
	codeInvalidDescription      httpx.Code = "INVALID_DESCRIPTION"
	codeInvalidCategory         httpx.Code = "INVALID_CATEGORY"
	codeInvalidReference        httpx.Code = "INVALID_REFERENCE"
	codeInvalidPriority         httpx.Code = "INVALID_PRIORITY"
	codeInvalidStatus           httpx.Code = "INVALID_STATUS"
	codeInvalidOrder            httpx.Code = "INVALID_ORDER"
	codeInvalidSort             httpx.Code = "INVALID_SORT"
	codeInvalidSpecialist       httpx.Code = "INVALID_SPECIALIST"
	codeSpecialistRequired      httpx.Code = "SPECIALIST_REQUIRED"
	codeTerminalTicket          httpx.Code = "TERMINAL_TICKET"
	codeForbidden               httpx.Code = "FORBIDDEN"
	codeInvalidStatusTransition httpx.Code = "INVALID_STATUS_TRANSITION"
)

func mapHTTPError(err error) error {
	var transitionErr *InvalidStatusTransitionError
	if errors.As(err, &transitionErr) {
		return httpx.Conflict(codeInvalidStatusTransition, transitionErr.Error())
	}

	switch {
	case errors.Is(err, ErrTicketNotFound):
		return httpx.NotFound(codeTicketNotFound, ErrTicketNotFound.Error())
	case errors.Is(err, ErrInvalidTitle):
		return httpx.BadRequest(codeInvalidTitle, ErrInvalidTitle.Error())
	case errors.Is(err, ErrInvalidDescription):
		return httpx.BadRequest(codeInvalidDescription, ErrInvalidDescription.Error())
	case errors.Is(err, ErrCategoryRequired):
		return httpx.BadRequest(codeInvalidCategory, ErrCategoryRequired.Error())
	case errors.Is(err, ErrInvalidCategory):
		return httpx.BadRequest(codeInvalidCategory, ErrInvalidCategory.Error())
	case errors.Is(err, ErrInvalidReference):
		return httpx.BadRequest(codeInvalidReference, ErrInvalidReference.Error())
	case errors.Is(err, ErrInvalidPriority):
		return httpx.BadRequest(codeInvalidPriority, ErrInvalidPriority.Error())
	case errors.Is(err, ErrInvalidStatus):
		return httpx.BadRequest(codeInvalidStatus, ErrInvalidStatus.Error())
	case errors.Is(err, ErrInvalidOrder):
		return httpx.BadRequest(codeInvalidOrder, ErrInvalidOrder.Error())
	case errors.Is(err, ErrInvalidSort):
		return httpx.BadRequest(codeInvalidSort, ErrInvalidSort.Error())
	case errors.Is(err, ErrInvalidSpecialist):
		return httpx.BadRequest(codeInvalidSpecialist, ErrInvalidSpecialist.Error())
	case errors.Is(err, ErrSpecialistRequired):
		return httpx.BadRequest(codeSpecialistRequired, ErrSpecialistRequired.Error())
	case errors.Is(err, ErrTerminalTicket):
		return httpx.Conflict(codeTerminalTicket, ErrTerminalTicket.Error())
	case errors.Is(err, ErrOnlyUsersCanCreate):
		return httpx.Forbidden(codeForbidden, ErrOnlyUsersCanCreate.Error())
	case errors.Is(err, ErrTicketCannotBeEdited):
		return httpx.Forbidden(codeForbidden, ErrTicketCannotBeEdited.Error())
	case errors.Is(err, ErrOnlyAdminsCanAssign):
		return httpx.Forbidden(codeForbidden, ErrOnlyAdminsCanAssign.Error())
	case errors.Is(err, ErrStatusChangeForbidden):
		return httpx.Forbidden(codeForbidden, ErrStatusChangeForbidden.Error())
	case errors.Is(err, ErrTicketAccessForbidden):
		return httpx.Forbidden(codeForbidden, ErrTicketAccessForbidden.Error())
	default:
		return comments.MapHTTPError(err)
	}
}
