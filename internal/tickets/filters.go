package tickets

import (
	"time"

	"github.com/google/uuid"
)

type Filter struct {
	Page        int
	PageSize    int
	Query       string
	Status      *Status
	Priority    *Priority
	CategoryID  *uuid.UUID
	AssignedTo  *uuid.UUID
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	CreatedBy   *uuid.UUID
	Sort        string
	Order       string
}

type Page struct {
	Items    []Record `json:"items"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
	Total    int64    `json:"total"`
}
