package analytics

import "context"

// Repository defines the read model contract for analytics queries.
type Repository interface {
	GetMonthlyReport(ctx context.Context, userID string, months int) (*MonthlyReport, error)
}
