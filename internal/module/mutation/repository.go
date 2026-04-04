package mutation

import "context"

// ListOptions contains mutation history query parameters.
type ListOptions struct {
	Page          int
	PageSize      int
	Type          string
	WalletID      string
	DebtID        string
	RelatedToDebt *bool
	From          string
	To            string
	SortBy        string
	SortDirection string
}

// ListResult contains paginated mutation history data.
type ListResult struct {
	Items      []Mutation
	TotalItems int
}

// Repository defines the persistence contract for mutation flows.
type Repository interface {
	Create(ctx context.Context, userID string, input UpsertInput) (*Mutation, error)
	List(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	GetByID(ctx context.Context, userID, mutationID string) (*Mutation, error)
	Update(ctx context.Context, userID, mutationID string, input UpsertInput) (*Mutation, error)
	Delete(ctx context.Context, userID, mutationID string) error
}
