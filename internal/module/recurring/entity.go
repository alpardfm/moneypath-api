package recurring

import "time"

// Rule is a recurring mutation template owned by a user.
type Rule struct {
	ID           string
	UserID       string
	WalletID     string
	CategoryID   *string
	Type         string
	Amount       string
	Description  string
	IntervalUnit string
	IntervalStep int
	StartAt      time.Time
	EndAt        *time.Time
	NextRunAt    time.Time
	LastRunAt    *time.Time
	IsActive     bool
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RunDueResult summarizes a recurring execution attempt.
type RunDueResult struct {
	Processed int              `json:"processed"`
	Skipped   []SkippedRunItem `json:"skipped"`
}

// SkippedRunItem describes one recurring rule that could not be executed.
type SkippedRunItem struct {
	RuleID string `json:"rule_id"`
	Reason string `json:"reason"`
}
