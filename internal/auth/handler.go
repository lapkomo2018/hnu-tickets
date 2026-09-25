package auth

import (
	"net/http"

	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/users"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// @Summary Register a user account
// @Description Creates a USER account and returns an access token. Clients cannot choose a role.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "Account details"
// @Success 201 {object} auth.SessionResponse
// @Failure 400,409 {object} httpx.ErrorEnvelope
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	var request RegisterRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	session, err := h.service.Register(r.Context(), RegisterInput(request))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toSessionResponse(session))
	return nil
}

// @Summary Create an access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Credentials"
// @Success 200 {object} auth.SessionResponse
// @Failure 400,401 {object} httpx.ErrorEnvelope
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var request LoginRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		return err
	}
	session, err := h.service.Login(r.Context(), LoginInput(request))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toSessionResponse(session))
	return nil
}

// @Summary Get the authenticated user
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} users.Response
// @Failure 401,404 {object} httpx.ErrorEnvelope
// @Router /auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) error {
	actor, err := appmiddleware.PrincipalFromContext(r.Context())
	if err != nil {
		return err
	}
	user, err := h.service.Me(r.Context(), actor.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, users.ToResponse(*user))
	return nil
}
