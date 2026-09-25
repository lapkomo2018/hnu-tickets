package analytics

import (
	"context"
	"time"
)

type Period struct {
	From *time.Time
	To   *time.Time
}

type Repository interface {
	Overview(context.Context, Period) (Overview, error)
	ByStatus(context.Context, Period) ([]Count, error)
	ByPriority(context.Context, Period) ([]Count, error)
	ByCategory(context.Context, Period) ([]CategoryCount, error)
	Specialists(context.Context, Period) ([]SpecialistWorkload, error)
}
