package leakage

import (
	"context"
	"testing"
)

type stubRepository struct {
	totalOutgoing    string
	categorySpends   []CategorySpend
	repeatedPatterns []RepeatedPattern
}

func (s stubRepository) GetTotalOutgoing(ctx context.Context, userID string, days int) (string, error) {
	return s.totalOutgoing, nil
}

func (s stubRepository) ListCategorySpends(ctx context.Context, userID string, days int) ([]CategorySpend, error) {
	return s.categorySpends, nil
}

func (s stubRepository) ListRepeatedPatterns(ctx context.Context, userID string, days int) ([]RepeatedPattern, error) {
	return s.repeatedPatterns, nil
}

func TestGetReportDetectsCategoryConcentrationAndRepeatedSpending(t *testing.T) {
	service := NewService(stubRepository{
		totalOutgoing: "1000.00",
		categorySpends: []CategorySpend{
			{CategoryID: "cat-1", CategoryName: "Food Delivery", TotalAmount: "520.00", TransactionCount: 8},
		},
		repeatedPatterns: []RepeatedPattern{
			{Description: "coffee", TotalAmount: "180.00", AverageAmount: "60.00", TransactionCount: 3},
		},
	})

	report, err := service.GetReport(context.Background(), "user-1", 30)
	if err != nil {
		t.Fatalf("GetReport() error = %v", err)
	}
	if len(report.Findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(report.Findings))
	}
	if len(report.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}
}

func TestGetReportRejectsInvalidDays(t *testing.T) {
	service := NewService(stubRepository{})

	if _, err := service.GetReport(context.Background(), "user-1", 3); err != ErrInvalidDays {
		t.Fatalf("expected ErrInvalidDays, got %v", err)
	}
}
