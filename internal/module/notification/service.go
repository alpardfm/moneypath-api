package notification

import (
	"context"
	"fmt"
	"time"
)

const upcomingRecurringWindow = 72 * time.Hour

// Service contains notification use cases.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService creates a notification service.
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

// GetReport returns the current notification feed for the authenticated user.
func (s *Service) GetReport(ctx context.Context, userID string) (*Report, error) {
	now := s.now()

	recurringItems, err := s.repo.ListDueRecurring(ctx, userID, now.Add(upcomingRecurringWindow))
	if err != nil {
		return nil, err
	}
	debtItems, err := s.repo.ListActiveDebtAlerts(ctx, userID, 5)
	if err != nil {
		return nil, err
	}

	report := &Report{}
	for _, item := range recurringItems {
		severity := "info"
		title := "Recurring rule is coming up"
		message := fmt.Sprintf("Recurring %s %s is scheduled at %s.", item.Type, item.Amount, item.NextRunAt.Format(time.RFC3339))
		if item.NextRunAt.Before(now) || item.NextRunAt.Equal(now) {
			severity = "warning"
			title = "Recurring rule is overdue"
			message = fmt.Sprintf("Recurring %s %s should have run at %s.", item.Type, item.Amount, item.NextRunAt.Format(time.RFC3339))
		}
		if item.Description != "" {
			message = fmt.Sprintf("%s Description: %s.", message, item.Description)
		}
		nextRunAt := item.NextRunAt
		report.Items = append(report.Items, Item{
			Type:         "recurring_due",
			Severity:     severity,
			Title:        title,
			Message:      message,
			ResourceID:   item.RuleID,
			ResourceType: "recurring_rule",
			OccurredAt:   &nextRunAt,
		})
	}

	for _, item := range debtItems {
		message := fmt.Sprintf("Debt %s still has remaining amount %s.", item.Name, item.RemainingAmount)
		if item.PaymentAmount != nil && *item.PaymentAmount != "" {
			message = fmt.Sprintf("%s Suggested payment amount is %s.", message, *item.PaymentAmount)
		}
		updatedAt := item.UpdatedAt
		report.Items = append(report.Items, Item{
			Type:         "debt_active",
			Severity:     "info",
			Title:        "Active debt still needs attention",
			Message:      message,
			ResourceID:   item.DebtID,
			ResourceType: "debt",
			OccurredAt:   &updatedAt,
		})
	}

	return report, nil
}
