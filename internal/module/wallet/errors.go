package wallet

import "errors"

var (
	ErrWalletNotFound        = errors.New("wallet not found")
	ErrWalletNameAlreadyUsed = errors.New("wallet name already used")
	ErrWalletBalanceNotZero  = errors.New("wallet balance must be zero")
	ErrWalletValidation      = errors.New("validation error")
)
