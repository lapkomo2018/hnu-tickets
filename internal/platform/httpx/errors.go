package httpx

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type Code string

const (
	CodeInternalError     Code = "INTERNAL_ERROR"
	CodeRouteNotFound     Code = "ROUTE_NOT_FOUND"
	CodeMethodNotAllowed  Code = "METHOD_NOT_ALLOWED"
	CodeInvalidJSON       Code = "INVALID_JSON"
	CodeInvalidID         Code = "INVALID_ID"
	CodeInvalidFilter     Code = "INVALID_FILTER"
	CodeInvalidSpecialist Code = "INVALID_SPECIALIST"
	CodeInvalidPeriod     Code = "INVALID_PERIOD"
	CodeUnauthorized      Code = "UNAUTHORIZED"
	CodeInvalidToken      Code = "INVALID_TOKEN"
)

type Error struct {
	status  int
	code    Code
	message string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Status() int { return e.status }

func (e *Error) Code() Code { return e.code }

func (e *Error) Message() string { return e.message }

func NewError(status int, code Code, message string) error {
	return &Error{status: status, code: code, message: message}
}

func WrapError(status int, code Code, message string, cause error) error {
	return &Error{status: status, code: code, message: message, cause: cause}
}

func BadRequest(code Code, message string) error {
	return NewError(http.StatusBadRequest, code, message)
}

func Unauthorized(code Code, message string) error {
	return NewError(http.StatusUnauthorized, code, message)
}

func Forbidden(code Code, message string) error {
	return NewError(http.StatusForbidden, code, message)
}

func NotFound(code Code, message string) error {
	return NewError(http.StatusNotFound, code, message)
}

func Conflict(code Code, message string) error {
	return NewError(http.StatusConflict, code, message)
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

type ErrorMapper func(error) error

type ErrorMiddleware struct {
	logger *slog.Logger
}

func NewErrorMiddleware(logger *slog.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{logger: logger}
}

func (m *ErrorMiddleware) HandleWith(mapper ErrorMapper) func(HandlerFunc) http.HandlerFunc {
	return func(next HandlerFunc) http.HandlerFunc {
		return m.Handle(func(w http.ResponseWriter, r *http.Request) error {
			if err := next(w, r); err != nil {
				return mapper(err)
			}
			return nil
		})
	}
}

func (m *ErrorMiddleware) Handle(next HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			m.Write(w, r, err)
		}
	}
}

func (m *ErrorMiddleware) Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				m.logger.ErrorContext(r.Context(), "panic recovered",
					"request_id", w.Header().Get("X-Request-ID"),
					"panic", recovered,
					"stack", string(debug.Stack()),
				)
				m.writeResponse(w, http.StatusInternalServerError, CodeInternalError, "an unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (m *ErrorMiddleware) NotFound(w http.ResponseWriter, r *http.Request) {
	m.Write(w, r, NewError(http.StatusNotFound, CodeRouteNotFound, "route not found"))
}

func (m *ErrorMiddleware) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	m.Write(w, r, NewError(http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed"))
}

func (m *ErrorMiddleware) Write(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message, internal := m.mapError(err)
	if internal {
		m.logger.ErrorContext(r.Context(), "request failed",
			"request_id", w.Header().Get("X-Request-ID"),
			"code", code,
			"error", err,
		)
	}
	m.writeResponse(w, status, code, message)
}

func (m *ErrorMiddleware) writeResponse(w http.ResponseWriter, status int, code Code, message string) {
	JSON(w, status, ErrorEnvelope{Error: ErrorResponse{Code: code, Message: message}})
}

func (m *ErrorMiddleware) mapError(err error) (int, Code, string, bool) {
	var transportErr *Error
	if errors.As(err, &transportErr) {
		return transportErr.status, transportErr.code, transportErr.message, transportErr.status >= http.StatusInternalServerError
	}
	return http.StatusInternalServerError, CodeInternalError, "an unexpected error occurred", true
}
