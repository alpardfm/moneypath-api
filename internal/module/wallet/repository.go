package wallet

import "context"

// ListOptions contains wallet list query parameters.
type ListOptions struct {
	Page     int
	PageSize int
}

// ListResult contains paginated wallet data.
type ListResult struct {
	Items      []Wallet
	TotalItems int
}

// Repository defines the persistence contract for wallet flows.
type Repository interface {
	Create(ctx context.Context, wallet *Wallet) error
	ListActive(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	ListArchived(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	GetByID(ctx context.Context, userID, walletID string) (*Wallet, error)
	UpdateName(ctx context.Context, userID, walletID, name string) (*Wallet, error)
	Inactivate(ctx context.Context, userID, walletID string) error
}
