package mutation

import "time"

// UpsertInput contains the payload for create and update mutation flows.
type UpsertInput struct {
	WalletID      string    `json:"wallet_id"`
	Type          string    `json:"type"`
	Amount        string    `json:"amount"`
	Description   string    `json:"description"`
	RelatedToDebt bool      `json:"related_to_debt"`
	HappenedAt    time.Time `json:"happened_at"`
}
