package summary

import (
	"context"
	"time"
)

// Service contains summary use cases.
type Service struct {
	repo Repository
}

// NewService creates a summary service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetReport returns the derived summary report for the authenticated user.
func (s *Service) GetReport(ctx context.Context, userID string, filter Filter) (*Report, error) {
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return nil, ErrInvalidPeriod
	}
	if filter.To != nil {
		end := filter.To.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.To = &end
	}
	return s.repo.GetReport(ctx, userID, filter)
}
