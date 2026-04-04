package profile

import (
	"context"
	"strings"

	"github.com/alpardfm/moneypath-api/internal/module/auth"
	"golang.org/x/crypto/bcrypt"
)

// Repository defines the persistence contract needed by profile flows.
type Repository interface {
	GetUserByID(ctx context.Context, userID string) (*auth.User, error)
	UpdateProfile(ctx context.Context, userID, email, username, fullName string) (*auth.User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}

// Service contains profile use cases.
type Service struct {
	repo Repository
}

// NewService creates a profile service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetMe returns the current user profile.
func (s *Service) GetMe(ctx context.Context, userID string) (*auth.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// UpdateMe updates the current user profile fields.
func (s *Service) UpdateMe(ctx context.Context, userID string, input UpdateProfileInput) (*auth.User, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	username := strings.TrimSpace(strings.ToLower(input.Username))
	fullName := strings.TrimSpace(input.FullName)
	if email == "" || username == "" || fullName == "" {
		return nil, auth.ErrValidation
	}

	return s.repo.UpdateProfile(ctx, userID, email, username, fullName)
}

// ChangePassword updates the current user password after verifying the old password.
func (s *Service) ChangePassword(ctx context.Context, userID string, input ChangePasswordInput) error {
	if strings.TrimSpace(input.CurrentPassword) == "" || strings.TrimSpace(input.NewPassword) == "" {
		return auth.ErrValidation
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)) != nil {
		return auth.ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, userID, string(hash))
}
