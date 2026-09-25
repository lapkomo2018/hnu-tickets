package auth

import "hnu-tickets/internal/users"

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type sessionResponse struct {
	AccessToken string         `json:"accessToken"`
	User        users.Response `json:"user"`
}

func toSessionResponse(session *Session) sessionResponse {
	return sessionResponse{AccessToken: session.AccessToken, User: users.ToResponse(session.User)}
}
