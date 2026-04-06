package dashboard

// WalletBalance represents one wallet balance row for the dashboard.
type WalletBalance struct {
	WalletID string `json:"wallet_id"`
	Name     string `json:"name"`
	Balance  string `json:"balance"`
}

// TrendPoint represents one month of incoming and outgoing movement for charts.
type TrendPoint struct {
	Month         string `json:"month"`
	TotalIncoming string `json:"total_incoming"`
	TotalOutgoing string `json:"total_outgoing"`
	NetFlow       string `json:"net_flow"`
}

// CategoryBreakdown represents one outgoing category aggregate for charts.
type CategoryBreakdown struct {
	CategoryID   string `json:"category_id,omitempty"`
	CategoryName string `json:"category_name"`
	TotalAmount  string `json:"total_amount"`
	Share        string `json:"share"`
}

// Overview contains the derived dashboard data.
type Overview struct {
	TotalAssets        string              `json:"total_assets"`
	TotalDebts         string              `json:"total_debts"`
	TotalIncoming      string              `json:"total_incoming"`
	TotalOutgoing      string              `json:"total_outgoing"`
	NetFlow            string              `json:"net_flow"`
	Wallets            []WalletBalance     `json:"wallets"`
	MonthlyTrend       []TrendPoint        `json:"monthly_trend"`
	OutgoingCategories []CategoryBreakdown `json:"outgoing_categories"`
}
