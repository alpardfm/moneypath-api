package debt

import "time"

// Debt is the master data for a user liability.
type Debt struct {
	ID              string
	UserID          string
	Name            string
	PrincipalAmount string
	RemainingAmount string
	TenorValue      *int
	TenorUnit       *string
	PaymentAmount   *string
	Status          string
	IsActive        bool
	Note            *string
	DeletedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
