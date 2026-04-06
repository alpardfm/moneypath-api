package leakage

import "context"

// Repository defines the read model contract for leakage detection queries.
type Repository interface {
	GetTotalOutgoing(ctx context.Context, userID string, days int) (string, error)
	ListCategorySpends(ctx context.Context, userID string, days int) ([]CategorySpend, error)
	ListRepeatedPatterns(ctx context.Context, userID string, days int) ([]RepeatedPattern, error)
}
