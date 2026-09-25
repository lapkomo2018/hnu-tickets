package analytics

import "errors"

var (
	ErrAdminRequired = errors.New("administrator access is required")
	ErrInvalidPeriod = errors.New("from must not be after to")
)
