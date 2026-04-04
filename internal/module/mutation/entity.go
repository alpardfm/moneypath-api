package mutation

import "time"

// Mutation is the source-of-truth financial event for wallet balance changes.
type Mutation struct {
	ID            string
	UserID        string
	WalletID      string
	DebtID        *string
	DebtAction    string
	Type          string
	Amount        string
	Description   string
	RelatedToDebt bool
	HappenedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
