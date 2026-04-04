package dashboard

// WalletBalance represents one wallet balance row for the dashboard.
type WalletBalance struct {
	WalletID string `json:"wallet_id"`
	Name     string `json:"name"`
	Balance  string `json:"balance"`
}

// Overview contains the derived dashboard data.
type Overview struct {
	TotalAssets   string          `json:"total_assets"`
	TotalDebts    string          `json:"total_debts"`
	TotalIncoming string          `json:"total_incoming"`
	TotalOutgoing string          `json:"total_outgoing"`
	NetFlow       string          `json:"net_flow"`
	Wallets       []WalletBalance `json:"wallets"`
}
