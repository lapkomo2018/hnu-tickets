package analytics

import "uuid"

type Overview struct {
	TotalTickets             int64   `json:"totalTickets"`
	OpenTickets              int64   `json:"openTickets"`
	InProgressTickets        int64   `json:"inProgressTickets"`
	ResolvedTickets          int64   `json:"resolvedTickets"`
	ClosedTickets            int64   `json:"closedTickets"`
	AverageResolutionSeconds float64 `json:"averageResolutionSeconds"`
}

type Count struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type CategoryCount struct {
	ID    uuid.UUID `json:"id" format:"uuid"`
	Name  string    `json:"name"`
	Count int64     `json:"count"`
}

type SpecialistWorkload struct {
	ID          uuid.UUID `json:"id" format:"uuid"`
	Name        string    `json:"name"`
	OpenTickets int64     `json:"openTickets"`
}
