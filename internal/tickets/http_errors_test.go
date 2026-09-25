package tickets

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/comments"
	"hnu-tickets/internal/platform/httpx"
)

func TestMapHTTPError(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   httpx.Code
	}{
		{ErrTicketNotFound, http.StatusNotFound, codeTicketNotFound},
		{ErrInvalidTitle, http.StatusBadRequest, codeInvalidTitle},
		{ErrInvalidDescription, http.StatusBadRequest, codeInvalidDescription},
		{ErrCategoryRequired, http.StatusBadRequest, codeInvalidCategory},
		{ErrInvalidCategory, http.StatusBadRequest, codeInvalidCategory},
		{ErrInvalidReference, http.StatusBadRequest, codeInvalidReference},
		{ErrInvalidPriority, http.StatusBadRequest, codeInvalidPriority},
		{ErrInvalidStatus, http.StatusBadRequest, codeInvalidStatus},
		{ErrInvalidOrder, http.StatusBadRequest, codeInvalidOrder},
		{ErrInvalidSort, http.StatusBadRequest, codeInvalidSort},
		{ErrInvalidSpecialist, http.StatusBadRequest, codeInvalidSpecialist},
		{ErrSpecialistRequired, http.StatusBadRequest, codeSpecialistRequired},
		{ErrTerminalTicket, http.StatusConflict, codeTerminalTicket},
		{ErrOnlyUsersCanCreate, http.StatusForbidden, codeForbidden},
		{ErrTicketCannotBeEdited, http.StatusForbidden, codeForbidden},
		{ErrOnlyAdminsCanAssign, http.StatusForbidden, codeForbidden},
		{ErrStatusChangeForbidden, http.StatusForbidden, codeForbidden},
		{ErrTicketAccessForbidden, http.StatusForbidden, codeForbidden},
		{&InvalidStatusTransitionError{From: StatusNew, To: StatusClosed}, http.StatusConflict, codeInvalidStatusTransition},
		{comments.ErrInvalidComment, http.StatusBadRequest, "INVALID_COMMENT"},
	}

	for _, test := range tests {
		mapped := mapHTTPError(test.err)
		var transportErr *httpx.Error
		require.ErrorAs(t, mapped, &transportErr)
		require.Equal(t, test.status, transportErr.Status())
		require.Equal(t, test.code, transportErr.Code())
	}
}
