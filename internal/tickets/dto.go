package tickets

import (
	"time"
	"uuid"
)

type CreateRequest struct {
	Title       string    `json:"title" validate:"required" minLength:"3" maxLength:"255" example:"Printer is not working"`
	Description string    `json:"description" validate:"required" minLength:"1" example:"Nothing prints after the driver update"`
	CategoryID  uuid.UUID `json:"categoryId" validate:"required" format:"uuid"`
	Priority    Priority  `json:"priority" validate:"required"`
}

type UpdateRequest struct {
	Title       *string    `json:"title" minLength:"3" maxLength:"255"`
	Description *string    `json:"description" minLength:"1"`
	CategoryID  *uuid.UUID `json:"categoryId" format:"uuid"`
	Priority    *Priority  `json:"priority"`
}

type AssignRequest struct {
	SpecialistID uuid.UUID `json:"specialistId" validate:"required" format:"uuid"`
}

type StatusRequest struct {
	Status Status `json:"status" validate:"required"`
}

type Response struct {
	ID          uuid.UUID       `json:"id" format:"uuid"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Status      Status          `json:"status"`
	Priority    Priority        `json:"priority"`
	Category    CategorySummary `json:"category"`
	CreatedBy   UserSummary     `json:"createdBy"`
	AssignedTo  *UserSummary    `json:"assignedTo"`
	CreatedAt   time.Time       `json:"createdAt" format:"date-time"`
	UpdatedAt   time.Time       `json:"updatedAt" format:"date-time"`
	ResolvedAt  *time.Time      `json:"resolvedAt" format:"date-time"`
	ClosedAt    *time.Time      `json:"closedAt" format:"date-time"`
}

type PageResponse struct {
	Items    []Response `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int64      `json:"total"`
}

func toResponse(record Record) Response {
	var assignedTo *UserSummary
	if record.AssignedTo != nil && record.AssigneeName != nil {
		assignedTo = &UserSummary{ID: *record.AssignedTo, Name: *record.AssigneeName}
	}
	return Response{
		ID: record.ID, Title: record.Title, Description: record.Description,
		Status: record.Status, Priority: record.Priority, Category: record.Category,
		CreatedBy: record.Creator, AssignedTo: assignedTo, CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt, ResolvedAt: record.ResolvedAt, ClosedAt: record.ClosedAt,
	}
}
