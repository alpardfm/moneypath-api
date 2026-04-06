package leakage

// CategorySpend represents one outgoing category aggregate in the selected period.
type CategorySpend struct {
	CategoryID       string `json:"category_id,omitempty"`
	CategoryName     string `json:"category_name"`
	TotalAmount      string `json:"total_amount"`
	TransactionCount int    `json:"transaction_count"`
}

// RepeatedPattern represents one repeated outgoing description pattern.
type RepeatedPattern struct {
	Description      string `json:"description"`
	TotalAmount      string `json:"total_amount"`
	AverageAmount    string `json:"average_amount"`
	TransactionCount int    `json:"transaction_count"`
}

// Finding represents one leakage signal detected from recent outgoing activity.
type Finding struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Amount      string `json:"amount"`
	Share       string `json:"share,omitempty"`
	Occurrences int    `json:"occurrences,omitempty"`
}

// Report contains the leakage detection output for the selected period.
type Report struct {
	Days             int               `json:"days"`
	TotalOutgoing    string            `json:"total_outgoing"`
	CategorySpends   []CategorySpend   `json:"category_spends"`
	RepeatedPatterns []RepeatedPattern `json:"repeated_patterns"`
	Findings         []Finding         `json:"findings"`
	Recommendations  []string          `json:"recommendations"`
}
