package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"hnu-tickets/internal/platform/apperror"
)

const MaxBodyBytes = 1 << 20

type errorEnvelope struct {
	Error errorResponse `json:"error"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return apperror.Wrap(apperror.KindValidation, "INVALID_JSON", "request body is not valid JSON", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return apperror.New(apperror.KindValidation, "INVALID_JSON", "request body must contain one JSON value")
	}
	return nil
}

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(w).Encode(value)
	}
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		slog.ErrorContext(r.Context(), "request failed", "error", err)
		JSON(w, http.StatusInternalServerError, errorEnvelope{Error: errorResponse{
			Code: "INTERNAL_ERROR", Message: "an unexpected error occurred",
		}})
		return
	}

	status := statusForKind(appErr.Kind)
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "request failed", "code", appErr.Code, "error", appErr.Cause)
	}
	JSON(w, status, errorEnvelope{Error: errorResponse{Code: appErr.Code, Message: appErr.Message}})
}

func statusForKind(kind apperror.Kind) int {
	switch kind {
	case apperror.KindValidation:
		return http.StatusBadRequest
	case apperror.KindUnauthorized:
		return http.StatusUnauthorized
	case apperror.KindForbidden:
		return http.StatusForbidden
	case apperror.KindNotFound:
		return http.StatusNotFound
	case apperror.KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
