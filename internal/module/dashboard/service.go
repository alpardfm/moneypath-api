package dashboard

import "context"

// Service contains dashboard use cases.
type Service struct {
	repo Repository
}

// NewService creates a dashboard service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetOverview returns the dashboard overview for the authenticated user.
func (s *Service) GetOverview(ctx context.Context, userID string) (*Overview, error) {
	return s.repo.GetOverview(ctx, userID)
}
