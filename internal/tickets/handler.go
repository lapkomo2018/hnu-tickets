package tickets

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	var request CreateRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	record, err := h.service.Create(r.Context(), actor, CreateInput(request))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toResponse(*record))
	return nil
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) error {
	id, actor, err := requestContext(r)
	if err != nil {
		return err
	}
	record, err := h.service.Get(r.Context(), actor, id)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toResponse(*record))
	return nil
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	filter, err := parseFilter(r)
	if err != nil {
		return err
	}
	page, err := h.service.List(r.Context(), actor, filter)
	if err != nil {
		return err
	}
	items := make([]Response, 0, len(page.Items))
	for _, record := range page.Items {
		items = append(items, toResponse(record))
	}
	httpx.JSON(w, http.StatusOK, PageResponse{Items: items, Page: page.Page, PageSize: page.PageSize, Total: page.Total})
	return nil
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	id, actor, err := requestContext(r)
	if err != nil {
		return err
	}
	var request UpdateRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	record, err := h.service.Update(r.Context(), actor, id, UpdateInput(request))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toResponse(*record))
	return nil
}

func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) error {
	id, actor, err := requestContext(r)
	if err != nil {
		return err
	}
	var request AssignRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	if request.SpecialistID == uuid.Nil {
		return httpx.NewError(http.StatusBadRequest, httpx.CodeInvalidSpecialist, "specialistId is required")
	}
	record, err := h.service.Assign(r.Context(), actor, id, request.SpecialistID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toResponse(*record))
	return nil
}

func (h *Handler) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	id, actor, err := requestContext(r)
	if err != nil {
		return err
	}
	var request StatusRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	record, err := h.service.ChangeStatus(r.Context(), actor, id, request.Status)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toResponse(*record))
	return nil
}

func requestContext(r *http.Request) (uuid.UUID, appmiddleware.Principal, error) {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return uuid.Nil, appmiddleware.Principal{}, err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return uuid.Nil, appmiddleware.Principal{}, err
	}
	return id, actor, nil
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
	return httpx.NewError(http.StatusBadRequest, httpx.CodeInvalidFilter, name+" filter is invalid")
}
