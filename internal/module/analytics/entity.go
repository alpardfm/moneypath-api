package analytics

// MonthlyPoint represents one month of aggregated financial movement.
type MonthlyPoint struct {
	Month         string `json:"month"`
	TotalIncoming string `json:"total_incoming"`
	TotalOutgoing string `json:"total_outgoing"`
	NetFlow       string `json:"net_flow"`
}

// MonthlyReport contains the monthly analytics series.
type MonthlyReport struct {
	Months int            `json:"months"`
	Items  []MonthlyPoint `json:"items"`
}
