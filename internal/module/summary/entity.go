package summary

import "time"

// Filter defines the summary period filter.
type Filter struct {
	From *time.Time
	To   *time.Time
}

// WalletBalance represents one wallet balance row for the summary response.
type WalletBalance struct {
	WalletID string `json:"wallet_id"`
	Name     string `json:"name"`
	Balance  string `json:"balance"`
}

// Report contains the derived summary data.
type Report struct {
	TotalAssets   string          `json:"total_assets"`
	TotalDebts    string          `json:"total_debts"`
	TotalIncoming string          `json:"total_incoming"`
	TotalOutgoing string          `json:"total_outgoing"`
	NetFlow       string          `json:"net_flow"`
	Wallets       []WalletBalance `json:"wallets"`
	From          *time.Time      `json:"from,omitempty"`
	To            *time.Time      `json:"to,omitempty"`
}
