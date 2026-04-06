package export

import "context"

// Repository defines the export read model contract.
type Repository interface {
	ListMutations(ctx context.Context, userID string, filter MutationFilter) ([]MutationRow, error)
}
