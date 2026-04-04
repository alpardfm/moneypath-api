package mutation

import "errors"

var (
	ErrMutationNotFound          = errors.New("mutation not found")
	ErrMutationValidation        = errors.New("validation error")
	ErrMutationDeleteNotAllowed  = errors.New("mutation delete is not allowed")
	ErrInsufficientWalletBalance = errors.New("insufficient wallet balance")
	ErrMutationWalletNotFound    = errors.New("wallet not found")
	ErrMutationDebtNotFound      = errors.New("debt not found")
	ErrInvalidDebtRelation       = errors.New("invalid debt relation")
	ErrDebtStateChanged          = errors.New("debt state changed by later operations")
)
