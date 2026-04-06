package analytics

import (
	"context"
	"testing"
)

type stubRepository struct {
	getMonthlyReportFn func(ctx context.Context, userID string, months int) (*MonthlyReport, error)
}

func (s *stubRepository) GetMonthlyReport(ctx context.Context, userID string, months int) (*MonthlyReport, error) {
	return s.getMonthlyReportFn(ctx, userID, months)
}

func TestGetMonthlyReportUsesDefaultMonths(t *testing.T) {
	service := NewService(&stubRepository{
		getMonthlyReportFn: func(ctx context.Context, userID string, months int) (*MonthlyReport, error) {
			if months != 6 {
				t.Fatalf("expected default months=6, got %d", months)
			}
			return &MonthlyReport{Months: months}, nil
		},
	})

	if _, err := service.GetMonthlyReport(context.Background(), "user-1", 0); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGetMonthlyReportRejectsInvalidMonths(t *testing.T) {
	service := NewService(&stubRepository{})

	if _, err := service.GetMonthlyReport(context.Background(), "user-1", 25); err != ErrInvalidMonths {
		t.Fatalf("expected ErrInvalidMonths, got %v", err)
	}
}
