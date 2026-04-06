package notification

import (
	"context"
	"time"
)

// RecurringAlert represents one due recurring rule.
type RecurringAlert struct {
	RuleID      string
	Description string
	Amount      string
	Type        string
	NextRunAt   time.Time
}

// DebtAlert represents one active debt reminder candidate.
type DebtAlert struct {
	DebtID          string
	Name            string
	RemainingAmount string
	PaymentAmount   *string
	UpdatedAt       time.Time
}

// Repository defines the read model contract for notification queries.
type Repository interface {
	ListDueRecurring(ctx context.Context, userID string, until time.Time) ([]RecurringAlert, error)
	ListActiveDebtAlerts(ctx context.Context, userID string, limit int) ([]DebtAlert, error)
}
