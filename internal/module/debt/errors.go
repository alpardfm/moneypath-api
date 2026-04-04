package debt

import "errors"

var (
	ErrDebtNotFound         = errors.New("debt not found")
	ErrDebtValidation       = errors.New("validation error")
	ErrDebtRemainingNotZero = errors.New("debt remaining amount must be zero")
)
