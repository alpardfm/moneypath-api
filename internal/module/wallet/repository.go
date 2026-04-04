package wallet

import "context"

// Repository defines the persistence contract for wallet flows.
type Repository interface {
	Create(ctx context.Context, wallet *Wallet) error
	ListActive(ctx context.Context, userID string) ([]Wallet, error)
	GetByID(ctx context.Context, userID, walletID string) (*Wallet, error)
	UpdateName(ctx context.Context, userID, walletID, name string) (*Wallet, error)
	Inactivate(ctx context.Context, userID, walletID string) error
}
