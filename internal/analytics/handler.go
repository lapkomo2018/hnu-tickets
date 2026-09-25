package analytics

import (
	"net/http"
	"time"

	"hnu-tickets/internal/platform/apperror"
	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	actor, period, ok := requestInput(w, r)
	if !ok {
		return
	}
	result, err := h.service.Overview(r.Context(), actor, period)
	writeResult(w, r, result, err)
}

func (h *Handler) Statuses(w http.ResponseWriter, r *http.Request) {
	actor, period, ok := requestInput(w, r)
	if !ok {
		return
	}
	result, err := h.service.ByStatus(r.Context(), actor, period)
	writeResult(w, r, result, err)
}

func (h *Handler) Priorities(w http.ResponseWriter, r *http.Request) {
	actor, period, ok := requestInput(w, r)
	if !ok {
		return
	}
	result, err := h.service.ByPriority(r.Context(), actor, period)
	writeResult(w, r, result, err)
}

func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	actor, period, ok := requestInput(w, r)
	if !ok {
		return
	}
	result, err := h.service.ByCategory(r.Context(), actor, period)
	writeResult(w, r, result, err)
}

func (h *Handler) Specialists(w http.ResponseWriter, r *http.Request) {
	actor, period, ok := requestInput(w, r)
	if !ok {
		return
	}
	result, err := h.service.Specialists(r.Context(), actor, period)
	writeResult(w, r, result, err)
}

func requestInput(w http.ResponseWriter, r *http.Request) (appmiddleware.Principal, Period, bool) {
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return appmiddleware.Principal{}, Period{}, false
	}
	period := Period{}
	if value := r.URL.Query().Get("from"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			httpx.Error(w, r, apperror.New(apperror.KindValidation, "INVALID_PERIOD", "from must be an RFC3339 timestamp"))
			return appmiddleware.Principal{}, Period{}, false
		}
		period.From = &parsed
	}
	if value := r.URL.Query().Get("to"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			httpx.Error(w, r, apperror.New(apperror.KindValidation, "INVALID_PERIOD", "to must be an RFC3339 timestamp"))
			return appmiddleware.Principal{}, Period{}, false
		}
		period.To = &parsed
	}
	return actor, period, true
}

func writeResult(w http.ResponseWriter, r *http.Request, result any, err error) {
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, result)
}
