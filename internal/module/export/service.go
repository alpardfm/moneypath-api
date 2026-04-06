package export

import (
	"context"
	"strings"
	"time"
)

// Service contains export use cases.
type Service struct {
	repo Repository
}

// NewService creates an export service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ExportMutations returns rows for CSV export.
func (s *Service) ExportMutations(ctx context.Context, userID string, filter MutationFilter) ([]MutationRow, error) {
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return nil, ErrInvalidFilter
	}
	if filter.To != nil {
		end := filter.To.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.To = &end
	}
	filter.Type = strings.TrimSpace(filter.Type)
	filter.WalletID = strings.TrimSpace(filter.WalletID)
	filter.CategoryID = strings.TrimSpace(filter.CategoryID)
	filter.DebtID = strings.TrimSpace(filter.DebtID)

	if filter.Type != "" && filter.Type != "masuk" && filter.Type != "keluar" {
		return nil, ErrInvalidFilter
	}

	return s.repo.ListMutations(ctx, userID, filter)
}
