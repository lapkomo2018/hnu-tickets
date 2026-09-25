package categories

import (
	"net/http"

	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// @Summary List categories
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Success 200 {array} categories.Response
// @Failure 401 {object} httpx.ErrorEnvelope
// @Router /categories [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	categories, err := h.service.List(r.Context())
	if err != nil {
		return err
	}
	response := make([]Response, 0, len(categories))
	for _, category := range categories {
		response = append(response, toResponse(category))
	}
	httpx.JSON(w, http.StatusOK, response)
	return nil
}

// @Summary Create a category
// @Description ADMIN only.
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body categories.Request true "Category details"
// @Success 201 {object} categories.Response
// @Failure 400,401,403,409 {object} httpx.ErrorEnvelope
// @Router /categories [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	var requestBody Request
	if err := httpx.DecodeJSON(w, r, &requestBody); err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	category, err := h.service.Create(r.Context(), actor, Input(requestBody))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toResponse(*category))
	return nil
}

// @Summary Update a category
// @Description ADMIN only.
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category identifier" format(uuid)
// @Param request body categories.Request true "Category details"
// @Success 200 {object} categories.Response
// @Failure 400,401,403,404,409 {object} httpx.ErrorEnvelope
// @Router /categories/{id} [patch]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	var requestBody Request
	if err := httpx.DecodeJSON(w, r, &requestBody); err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	category, err := h.service.Update(r.Context(), actor, id, Input(requestBody))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toResponse(*category))
	return nil
}

// @Summary Delete an unused category
// @Description ADMIN only. Categories still referenced by tickets cannot be deleted.
// @Tags Categories
// @Security BearerAuth
// @Param id path string true "Category identifier" format(uuid)
// @Success 204 "Category deleted"
// @Failure 400,401,403,404,409 {object} httpx.ErrorEnvelope
// @Router /categories/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.UUIDParam(r, "id")
	if err != nil {
		return err
	}
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	if err := h.service.Delete(r.Context(), actor, id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
