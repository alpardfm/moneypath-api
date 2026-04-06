package recurring

import "errors"

var (
	ErrRuleNotFound         = errors.New("recurring rule not found")
	ErrRuleValidation       = errors.New("validation error")
	ErrRuleWalletNotFound   = errors.New("wallet not found")
	ErrRuleCategoryNotFound = errors.New("category not found")
	ErrRuleCategoryMismatch = errors.New("category type must match recurring type")
)
