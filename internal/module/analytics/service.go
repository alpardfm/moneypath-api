package analytics

import "context"

// Service contains analytics use cases.
type Service struct {
	repo Repository
}

// NewService creates an analytics service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetMonthlyReport returns monthly aggregated mutation data.
func (s *Service) GetMonthlyReport(ctx context.Context, userID string, months int) (*MonthlyReport, error) {
	if months == 0 {
		months = 6
	}
	if months < 1 || months > 24 {
		return nil, ErrInvalidMonths
	}
	return s.repo.GetMonthlyReport(ctx, userID, months)
}
