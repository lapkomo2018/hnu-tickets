package categories

import "uuid"

type Category struct {
	ID          uuid.UUID
	Name        string
	Description string
}
