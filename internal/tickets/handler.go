package tickets

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.Error(w, r, err)
		return
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	record, err := h.service.Create(r.Context(), actor, CreateInput(request))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toResponse(*record))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, actor, ok := requestContext(w, r)
	if !ok {
		return
	}
	record, err := h.service.Get(r.Context(), actor, id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toResponse(*record))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	filter, err := parseFilter(r)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	page, err := h.service.List(r.Context(), actor, filter)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	items := make([]Response, 0, len(page.Items))
	for _, record := range page.Items {
		items = append(items, toResponse(record))
	}
	httpx.JSON(w, http.StatusOK, pageResponse{Items: items, Page: page.Page, PageSize: page.PageSize, Total: page.Total})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, actor, ok := requestContext(w, r)
	if !ok {
		return
	}
	var request updateRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.Error(w, r, err)
		return
	}
	record, err := h.service.Update(r.Context(), actor, id, UpdateInput(request))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toResponse(*record))
}

func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	id, actor, ok := requestContext(w, r)
	if !ok {
		return
	}
	var request assignRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.Error(w, r, err)
		return
	}
	if request.SpecialistID == uuid.Nil {
		httpx.Error(w, r, apperror.New(apperror.KindValidation, "INVALID_SPECIALIST", "specialistId is required"))
		return
	}
	record, err := h.service.Assign(r.Context(), actor, id, request.SpecialistID)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toResponse(*record))
}

func (h *Handler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	id, actor, ok := requestContext(w, r)
	if !ok {
		return
	}
	var request statusRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.Error(w, r, err)
		return
	}
	record, err := h.service.ChangeStatus(r.Context(), actor, id, request.Status)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toResponse(*record))
}

func requestContext(w http.ResponseWriter, r *http.Request) (uuid.UUID, appmiddleware.Principal, bool) {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return uuid.Nil, appmiddleware.Principal{}, false
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return uuid.Nil, appmiddleware.Principal{}, false
	}
	return id, actor, true
}

func parseFilter(r *http.Request) (Filter, error) {
	query := r.URL.Query()
	filter := Filter{Query: query.Get("q"), Sort: query.Get("sort"), Order: query.Get("order")}
	var err error
	if value := query.Get("page"); value != "" {
		filter.Page, err = strconv.Atoi(value)
		if err != nil {
			return Filter{}, invalidFilter("page")
		}
	}
	if value := query.Get("pageSize"); value != "" {
		filter.PageSize, err = strconv.Atoi(value)
		if err != nil {
			return Filter{}, invalidFilter("pageSize")
		}
	}
	if value := query.Get("status"); value != "" {
		status := Status(value)
		filter.Status = &status
	}
	if value := query.Get("priority"); value != "" {
		priority := Priority(value)
		filter.Priority = &priority
	}
	if filter.CategoryID, err = parseOptionalUUID(query.Get("categoryId")); err != nil {
		return Filter{}, invalidFilter("categoryId")
	}
	if filter.AssignedTo, err = parseOptionalUUID(query.Get("assignedTo")); err != nil {
		return Filter{}, invalidFilter("assignedTo")
	}
	if filter.CreatedFrom, err = parseOptionalTime(query.Get("createdFrom")); err != nil {
		return Filter{}, invalidFilter("createdFrom")
	}
	if filter.CreatedTo, err = parseOptionalTime(query.Get("createdTo")); err != nil {
		return Filter{}, invalidFilter("createdTo")
	}
	return filter, nil
}

func parseOptionalUUID(value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value)
	return &parsed, err
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return &parsed, err
}

func invalidFilter(name string) error {
	return apperror.New(apperror.KindValidation, "INVALID_FILTER", name+" filter is invalid")
}
