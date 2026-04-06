package healthscore

import (
	"context"
	"testing"
)

type stubRepository struct {
	snapshot *Snapshot
}

func (s stubRepository) GetSnapshot(ctx context.Context, userID string) (*Snapshot, error) {
	return s.snapshot, nil
}

func TestGetReportStrongScore(t *testing.T) {
	service := NewService(stubRepository{
		snapshot: &Snapshot{
			TotalAssets:    "12000.00",
			TotalDebts:     "1000.00",
			RecentIncoming: "9000.00",
			RecentOutgoing: "3000.00",
		},
	})

	report, err := service.GetReport(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetReport() error = %v", err)
	}
	if report.Score < 80 {
		t.Fatalf("expected strong score, got %d", report.Score)
	}
	if report.Status != "strong" {
		t.Fatalf("expected strong status, got %q", report.Status)
	}
}

func TestGetReportRiskScore(t *testing.T) {
	service := NewService(stubRepository{
		snapshot: &Snapshot{
			TotalAssets:    "100.00",
			TotalDebts:     "1200.00",
			RecentIncoming: "300.00",
			RecentOutgoing: "1800.00",
		},
	})

	report, err := service.GetReport(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetReport() error = %v", err)
	}
	if report.Score >= 40 {
		t.Fatalf("expected risk score, got %d", report.Score)
	}
	if report.Status != "risk" {
		t.Fatalf("expected risk status, got %q", report.Status)
	}
	if len(report.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}
}
