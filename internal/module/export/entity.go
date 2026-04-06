package export

import "time"

// MutationFilter defines the supported export filters.
type MutationFilter struct {
	From       *time.Time
	To         *time.Time
	Type       string
	WalletID   string
	CategoryID string
	DebtID     string
}

// MutationRow represents one exported mutation row.
type MutationRow struct {
	ID            string
	WalletName    string
	CategoryName  string
	DebtName      string
	Type          string
	Amount        string
	Description   string
	RelatedToDebt bool
	HappenedAt    time.Time
	CreatedAt     time.Time
}
