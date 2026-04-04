package debt

import "context"

// ListOptions contains debt list query parameters.
type ListOptions struct {
	Page     int
	PageSize int
}

// ListResult contains paginated debt data.
type ListResult struct {
	Items      []Debt
	TotalItems int
}

// Repository defines the persistence contract for debt flows.
type Repository interface {
	Create(ctx context.Context, debt *Debt) error
	List(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	GetByID(ctx context.Context, userID, debtID string) (*Debt, error)
	Update(ctx context.Context, debt *Debt) (*Debt, error)
	Inactivate(ctx context.Context, userID, debtID string) error
}
