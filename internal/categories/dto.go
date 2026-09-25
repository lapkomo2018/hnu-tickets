package categories

import "github.com/google/uuid"

type Request struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Response struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func toResponse(category Category) Response {
	return Response{ID: category.ID, Name: category.Name, Description: category.Description}
}
