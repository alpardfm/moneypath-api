package dashboard

import "context"

// Repository defines the read model contract for dashboard queries.
type Repository interface {
	GetOverview(ctx context.Context, userID string) (*Overview, error)
}
