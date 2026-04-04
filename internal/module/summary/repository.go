package summary

import "context"

// Repository defines the read model contract for summary queries.
type Repository interface {
	GetReport(ctx context.Context, userID string, filter Filter) (*Report, error)
}
