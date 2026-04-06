package settings

import (
	"context"
	"strings"

	"github.com/alpardfm/moneypath-api/internal/module/auth"
)

// Repository defines the persistence contract needed by settings flows.
type Repository interface {
	GetUserByID(ctx context.Context, userID string) (*auth.User, error)
	UpdateSettings(ctx context.Context, userID, preferredCurrency, timezone, dateFormat, weekStartDay string) (*auth.User, error)
}

// Service contains settings use cases.
type Service struct {
	repo Repository
}

// NewService creates a settings service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Get returns the current user settings.
func (s *Service) Get(ctx context.Context, userID string) (*auth.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// Update updates the current user settings.
func (s *Service) Update(ctx context.Context, userID string, input UpdateInput) (*auth.User, error) {
	preferredCurrency := strings.ToUpper(strings.TrimSpace(input.PreferredCurrency))
	timezone := strings.TrimSpace(input.Timezone)
	dateFormat := strings.TrimSpace(input.DateFormat)
	weekStartDay := strings.ToLower(strings.TrimSpace(input.WeekStartDay))

	if preferredCurrency == "" || timezone == "" || dateFormat == "" || weekStartDay == "" {
		return nil, auth.ErrValidation
	}
	if dateFormat != "YYYY-MM-DD" && dateFormat != "DD-MM-YYYY" && dateFormat != "MM-DD-YYYY" {
		return nil, auth.ErrValidation
	}
	if weekStartDay != "monday" && weekStartDay != "sunday" {
		return nil, auth.ErrValidation
	}

	return s.repo.UpdateSettings(ctx, userID, preferredCurrency, timezone, dateFormat, weekStartDay)
}
