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
			return &Overview{TotalAssets: "100.00"}, nil
		},
	})

	result, err := service.GetOverview(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetOverview() error = %v", err)
	}
	if result.TotalAssets != "100.00" {
		t.Fatalf("expected total assets 100.00, got %q", result.TotalAssets)
	}
}
