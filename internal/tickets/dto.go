package tickets

import (
	"time"

	"github.com/google/uuid"
)

type createRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CategoryID  uuid.UUID `json:"categoryId"`
	Priority    Priority  `json:"priority"`
}

type updateRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	CategoryID  *uuid.UUID `json:"categoryId"`
	Priority    *Priority  `json:"priority"`
}

type assignRequest struct {
	SpecialistID uuid.UUID `json:"specialistId"`
}

type statusRequest struct {
	Status Status `json:"status"`
}

type Response struct {
	ID          uuid.UUID       `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Status      Status          `json:"status"`
	Priority    Priority        `json:"priority"`
	Category    CategorySummary `json:"category"`
	CreatedBy   UserSummary     `json:"createdBy"`
	AssignedTo  *UserSummary    `json:"assignedTo"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	ResolvedAt  *time.Time      `json:"resolvedAt"`
	ClosedAt    *time.Time      `json:"closedAt"`
}

type pageResponse struct {
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
