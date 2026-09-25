package categories

import "errors"

var (
	ErrCategoryNotFound   = errors.New("category not found")
	ErrCategoryNameExists = errors.New("a category with this name already exists")
	ErrCategoryInUse      = errors.New("category is used by one or more tickets")
	ErrInvalidName        = errors.New("category name must contain 2 to 100 characters")
	ErrAdminRequired      = errors.New("administrator access is required")
)
