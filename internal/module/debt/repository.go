package debt

import "context"

// Repository defines the persistence contract for debt flows.
type Repository interface {
	Create(ctx context.Context, debt *Debt) error
	List(ctx context.Context, userID string) ([]Debt, error)
	GetByID(ctx context.Context, userID, debtID string) (*Debt, error)
	Update(ctx context.Context, debt *Debt) (*Debt, error)
	Inactivate(ctx context.Context, userID, debtID string) error
}
