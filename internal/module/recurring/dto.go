package recurring

import "time"

// CreateInput contains the create recurring rule payload.
type CreateInput struct {
	WalletID     string     `json:"wallet_id"`
	CategoryID   *string    `json:"category_id"`
	Type         string     `json:"type"`
	Amount       string     `json:"amount"`
	Description  string     `json:"description"`
	IntervalUnit string     `json:"interval_unit"`
	IntervalStep int        `json:"interval_step"`
	StartAt      time.Time  `json:"start_at"`
	EndAt        *time.Time `json:"end_at"`
}

// UpdateInput contains the update recurring rule payload.
type UpdateInput struct {
	WalletID     string     `json:"wallet_id"`
	CategoryID   *string    `json:"category_id"`
	Type         string     `json:"type"`
	Amount       string     `json:"amount"`
	Description  string     `json:"description"`
	IntervalUnit string     `json:"interval_unit"`
	IntervalStep int        `json:"interval_step"`
	StartAt      time.Time  `json:"start_at"`
	EndAt        *time.Time `json:"end_at"`
}
