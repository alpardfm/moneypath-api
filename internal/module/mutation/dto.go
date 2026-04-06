package mutation

import "time"

// UpsertInput contains the payload for create and update mutation flows.
type UpsertInput struct {
	WalletID      string        `json:"wallet_id"`
	CategoryID    *string       `json:"category_id"`
	DebtID        *string       `json:"debt_id"`
	Type          string        `json:"type"`
	Amount        string        `json:"amount"`
	Description   string        `json:"description"`
	RelatedToDebt bool          `json:"related_to_debt"`
	NewDebt       *NewDebtInput `json:"new_debt"`
	HappenedAt    time.Time     `json:"happened_at"`
}

// NewDebtInput contains the payload to create a new debt from an incoming mutation.
type NewDebtInput struct {
	Name          string  `json:"name"`
	Principal     string  `json:"principal_amount"`
	TenorValue    *int    `json:"tenor_value"`
	TenorUnit     *string `json:"tenor_unit"`
	PaymentAmount *string `json:"payment_amount"`
	Note          *string `json:"note"`
}
