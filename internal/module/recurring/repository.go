package recurring

import (
	"context"
	"time"
)

// ListOptions contains recurring list query parameters.
type ListOptions struct {
	Page     int
	PageSize int
	Type     string
}

// ListResult contains paginated recurring rule data.
type ListResult struct {
	Items      []Rule
	TotalItems int
}

// Repository defines the persistence contract for recurring flows.
type Repository interface {
	Create(ctx context.Context, rule *Rule) error
	List(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	GetByID(ctx context.Context, userID, ruleID string) (*Rule, error)
	Update(ctx context.Context, rule *Rule) (*Rule, error)
	Inactivate(ctx context.Context, userID, ruleID string) error
	RunDue(ctx context.Context, userID string, now time.Time) (*RunDueResult, error)
}
