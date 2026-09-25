package httpx

import (
	"encoding/json"
	"io"
	"net/http"
)

const MaxBodyBytes = 1 << 20

type ErrorEnvelope struct {
	Error ErrorResponse `json:"error"`
}

type ErrorResponse struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return WrapError(http.StatusBadRequest, CodeInvalidJSON, "request body is not valid JSON", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return NewError(http.StatusBadRequest, CodeInvalidJSON, "request body must contain one JSON value")
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
