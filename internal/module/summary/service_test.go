package summary

import (
	"context"
	"testing"
	"time"
)

type stubRepository struct {
	getFn func(ctx context.Context, userID string, filter Filter) (*Report, error)
}

func (s *stubRepository) GetReport(ctx context.Context, userID string, filter Filter) (*Report, error) {
	return s.getFn(ctx, userID, filter)
}

func TestGetReportRejectsInvalidPeriod(t *testing.T) {
	service := NewService(&stubRepository{})
	from := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	_, err := service.GetReport(context.Background(), "user-1", Filter{From: &from, To: &to})
	if err != ErrInvalidPeriod {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}
}
