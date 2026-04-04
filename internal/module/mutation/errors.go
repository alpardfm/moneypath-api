package mutation

import "errors"

var (
	ErrMutationNotFound          = errors.New("mutation not found")
	ErrMutationValidation        = errors.New("validation error")
	ErrMutationDeleteNotAllowed  = errors.New("mutation delete is not allowed")
	ErrInsufficientWalletBalance = errors.New("insufficient wallet balance")
	ErrMutationWalletNotFound    = errors.New("wallet not found")
	ErrDebtRelationNotSupported  = errors.New("debt-related mutation is not supported in this phase")
)
