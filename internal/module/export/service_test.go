package export

import (
	"context"
	"testing"
	"time"
)

type stubRepository struct {
	listMutationsFn func(ctx context.Context, userID string, filter MutationFilter) ([]MutationRow, error)
}

func (s *stubRepository) ListMutations(ctx context.Context, userID string, filter MutationFilter) ([]MutationRow, error) {
	return s.listMutationsFn(ctx, userID, filter)
}

func TestExportMutationsRejectsInvalidType(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.ExportMutations(context.Background(), "user-1", MutationFilter{Type: "other"})
	if err != ErrInvalidFilter {
		t.Fatalf("expected ErrInvalidFilter, got %v", err)
	}
}

func TestExportMutationsNormalizesToEndOfDay(t *testing.T) {
	service := NewService(&stubRepository{
		listMutationsFn: func(ctx context.Context, userID string, filter MutationFilter) ([]MutationRow, error) {
			if filter.To == nil || filter.To.Hour() != 23 {
				t.Fatalf("expected end-of-day filter, got %+v", filter.To)
			}
			return []MutationRow{}, nil
		},
	})

	to := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	if _, err := service.ExportMutations(context.Background(), "user-1", MutationFilter{To: &to}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
