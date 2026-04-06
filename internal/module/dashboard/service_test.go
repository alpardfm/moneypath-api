package dashboard

import (
	"context"
	"testing"
)

type stubRepository struct {
	getFn func(ctx context.Context, userID string) (*Overview, error)
}

func (s *stubRepository) GetOverview(ctx context.Context, userID string) (*Overview, error) {
	return s.getFn(ctx, userID)
}

func TestGetOverview(t *testing.T) {
	service := NewService(&stubRepository{
		getFn: func(ctx context.Context, userID string) (*Overview, error) {
			return &Overview{
				TotalAssets:        "100.00",
				MonthlyTrend:       []TrendPoint{{Month: "2026-04", TotalIncoming: "200.00"}},
				OutgoingCategories: []CategoryBreakdown{{CategoryName: "Food", TotalAmount: "50.00", Share: "25.00"}},
			}, nil
		},
	})

	result, err := service.GetOverview(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetOverview() error = %v", err)
	}
	if result.TotalAssets != "100.00" {
		t.Fatalf("expected total assets 100.00, got %q", result.TotalAssets)
	}
	if len(result.MonthlyTrend) != 1 {
		t.Fatalf("expected 1 trend point, got %d", len(result.MonthlyTrend))
	}
	if len(result.OutgoingCategories) != 1 {
		t.Fatalf("expected 1 category breakdown, got %d", len(result.OutgoingCategories))
	}
}
