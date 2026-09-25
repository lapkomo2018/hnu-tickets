package auth

import "hnu-tickets/internal/users"

type RegisterRequest struct {
	Name     string `json:"name" validate:"required" minLength:"2" maxLength:"100" example:"Maksym Lapko"`
	Email    string `json:"email" validate:"required" format:"email" example:"user@example.com"`
	Password string `json:"password" validate:"required" format:"password" minLength:"8" example:"password123"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required" format:"email" example:"user@example.com"`
	Password string `json:"password" validate:"required" format:"password" example:"password123"`
}

type SessionResponse struct {
	AccessToken string         `json:"accessToken"`
	User        users.Response `json:"user"`
}

func toSessionResponse(session *Session) SessionResponse {
	return SessionResponse{AccessToken: session.AccessToken, User: users.ToResponse(session.User)}
}
