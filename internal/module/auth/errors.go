package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailAlreadyUsed    = errors.New("email already used")
	ErrUsernameAlreadyUsed = errors.New("username already used")
	ErrInvalidToken        = errors.New("invalid token")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrUserNotFound        = errors.New("user not found")
	ErrValidation          = errors.New("validation error")
)
