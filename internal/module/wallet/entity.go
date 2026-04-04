package wallet

import "time"

// Wallet is the master data for a money container owned by a user.
type Wallet struct {
	ID        string
	UserID    string
	Name      string
	Balance   string
	IsActive  bool
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
