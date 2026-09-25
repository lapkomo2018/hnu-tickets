package auth

import "hnu-tickets/internal/users"

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SessionResponse struct {
	AccessToken string         `json:"accessToken"`
	User        users.Response `json:"user"`
}

func toSessionResponse(session *Session) SessionResponse {
	return SessionResponse{AccessToken: session.AccessToken, User: users.ToResponse(session.User)}
}
