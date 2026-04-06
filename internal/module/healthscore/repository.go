package healthscore

import "context"

// Snapshot contains raw aggregates fetched from storage.
type Snapshot struct {
	TotalAssets    string
	TotalDebts     string
	RecentIncoming string
	RecentOutgoing string
}

// Repository defines the read model contract for financial health scoring.
type Repository interface {
	GetSnapshot(ctx context.Context, userID string) (*Snapshot, error)
}
