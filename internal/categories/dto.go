package categories

import "github.com/google/uuid"

type Request struct {
	Name        string `json:"name" validate:"required" minLength:"2" maxLength:"100" example:"Hardware"`
	Description string `json:"description" example:"Printers, monitors and other devices"`
}

type Response struct {
	ID          uuid.UUID `json:"id" format:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func toResponse(category Category) Response {
	return Response{ID: category.ID, Name: category.Name, Description: category.Description}
}
