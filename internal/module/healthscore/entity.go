package healthscore

// Inputs contains the raw aggregates used to derive the score.
type Inputs struct {
	TotalAssets            string `json:"total_assets"`
	TotalDebts             string `json:"total_debts"`
	RecentIncoming         string `json:"recent_incoming"`
	RecentOutgoing         string `json:"recent_outgoing"`
	AverageMonthlyOutgoing string `json:"average_monthly_outgoing"`
}

// Metrics contains the intermediate ratios used by the scoring rules.
type Metrics struct {
	LiquidityMonths  string `json:"liquidity_months"`
	DebtToAssetRatio string `json:"debt_to_asset_ratio"`
	CashFlowCoverage string `json:"cash_flow_coverage"`
}

// Component contains one scoring component breakdown.
type Component struct {
	Name        string `json:"name"`
	Score       int    `json:"score"`
	MaxScore    int    `json:"max_score"`
	Description string `json:"description"`
}

// Report contains the derived financial health score.
type Report struct {
	Score           int         `json:"score"`
	Status          string      `json:"status"`
	Summary         string      `json:"summary"`
	Inputs          Inputs      `json:"inputs"`
	Metrics         Metrics     `json:"metrics"`
	Components      []Component `json:"components"`
	Recommendations []string    `json:"recommendations"`
}
