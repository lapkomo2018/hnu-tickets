package tickets

import "errors"

var (
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrTerminalTicket    = errors.New("ticket is closed or cancelled")
)
