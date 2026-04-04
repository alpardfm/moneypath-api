package auth

import "time"

// User is the authenticated ownership root in the application.
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	FullName     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Claims is the JWT payload used by the API.
type Claims struct {
	UserID string `json:"user_id"`
}
