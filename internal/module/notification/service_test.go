package notification

import (
	"context"
	"testing"
	"time"
)

type stubRepository struct {
	recurringItems []RecurringAlert
	debtItems      []DebtAlert
}

func (s stubRepository) ListDueRecurring(ctx context.Context, userID string, until time.Time) ([]RecurringAlert, error) {
	return s.recurringItems, nil
}

func (s stubRepository) ListActiveDebtAlerts(ctx context.Context, userID string, limit int) ([]DebtAlert, error) {
	return s.debtItems, nil
}

func TestGetReportBuildsRecurringAndDebtNotifications(t *testing.T) {
	now := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	service := NewService(stubRepository{
		recurringItems: []RecurringAlert{
			{RuleID: "rule-1", Amount: "100.00", Type: "keluar", NextRunAt: now.Add(-time.Hour)},
		},
		debtItems: []DebtAlert{
			{DebtID: "debt-1", Name: "Laptop", RemainingAmount: "500.00", UpdatedAt: now},
		},
	})
	service.now = func() time.Time { return now }

	report, err := service.GetReport(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetReport() error = %v", err)
	}
	if len(report.Items) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(report.Items))
	}
	if report.Items[0].Type != "recurring_due" {
		t.Fatalf("expected first item recurring_due, got %q", report.Items[0].Type)
	}
}
