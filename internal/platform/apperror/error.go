package apperror

import "fmt"

type Kind string

const (
	KindValidation   Kind = "validation"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindInternal     Kind = "internal"
)

type Error struct {
	Kind    Kind
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func New(kind Kind, code, message string) error {
	return &Error{Kind: kind, Code: code, Message: message}
}

func Wrap(kind Kind, code, message string, cause error) error {
	return &Error{Kind: kind, Code: code, Message: message, Cause: cause}
}
