package auth

import "context"

// Repository defines the persistence contract for auth and profile flows.
type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmailOrUsername(ctx context.Context, value string) (*User, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, userID, email, username, fullName string) (*User, error)
	UpdateSettings(ctx context.Context, userID, preferredCurrency, timezone, dateFormat, weekStartDay string) (*User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}
