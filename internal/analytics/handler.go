package analytics

import (
	"net/http"
	"time"

	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) error {
	actor, period, err := requestInput(r)
	if err != nil {
		return err
	}
	result, err := h.service.Overview(r.Context(), actor, period)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, result)
	return nil
}

func (h *Handler) Statuses(w http.ResponseWriter, r *http.Request) error {
	actor, period, err := requestInput(r)
	if err != nil {
		return err
	}
	result, err := h.service.ByStatus(r.Context(), actor, period)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, result)
	return nil
}

func (h *Handler) Priorities(w http.ResponseWriter, r *http.Request) error {
	actor, period, err := requestInput(r)
	if err != nil {
		return err
	}
	result, err := h.service.ByPriority(r.Context(), actor, period)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, result)
	return nil
}

func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) error {
	actor, period, err := requestInput(r)
	if err != nil {
		return err
	}
	result, err := h.service.ByCategory(r.Context(), actor, period)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, result)
	return nil
}

func (h *Handler) Specialists(w http.ResponseWriter, r *http.Request) error {
	actor, period, err := requestInput(r)
	if err != nil {
		return err
	}
	result, err := h.service.Specialists(r.Context(), actor, period)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, result)
	return nil
}

func requestInput(r *http.Request) (appmiddleware.Principal, Period, error) {
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return appmiddleware.Principal{}, Period{}, err
	}
	period := Period{}
	if value := r.URL.Query().Get("from"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return appmiddleware.Principal{}, Period{}, httpx.NewError(http.StatusBadRequest, httpx.CodeInvalidPeriod, "from must be an RFC3339 timestamp")
		}
		period.From = &parsed
	}
	if value := r.URL.Query().Get("to"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return appmiddleware.Principal{}, Period{}, httpx.NewError(http.StatusBadRequest, httpx.CodeInvalidPeriod, "to must be an RFC3339 timestamp")
		}
		period.To = &parsed
	}
	return actor, period, nil
}
