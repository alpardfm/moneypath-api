package category

import "errors"

var (
	ErrCategoryNotFound        = errors.New("category not found")
	ErrCategoryNameAlreadyUsed = errors.New("category name already used")
	ErrCategoryValidation      = errors.New("validation error")
)
