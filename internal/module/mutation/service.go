package mutation

import (
	"context"
	"strings"
)

// Service contains mutation use cases.
type Service struct {
	repo Repository
}

// NewService creates a mutation service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a mutation and updates wallet balance consistently.
func (s *Service) Create(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, userID, sanitizeInput(input))
}

// List returns mutation history for the authenticated user.
func (s *Service) List(ctx context.Context, userID string) ([]Mutation, error) {
	return s.repo.List(ctx, userID)
}

// GetByID returns one mutation owned by the authenticated user.
func (s *Service) GetByID(ctx context.Context, userID, mutationID string) (*Mutation, error) {
	return s.repo.GetByID(ctx, userID, mutationID)
}

// Update edits a mutation using rollback-and-reapply logic.
func (s *Service) Update(ctx context.Context, userID, mutationID string, input UpsertInput) (*Mutation, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, userID, mutationID, sanitizeInput(input))
}

// Delete is not allowed for mutation history.
func (s *Service) Delete(ctx context.Context, userID, mutationID string) error {
	return s.repo.Delete(ctx, userID, mutationID)
}

func validateInput(input UpsertInput) error {
	if strings.TrimSpace(input.WalletID) == "" ||
		strings.TrimSpace(input.Amount) == "" ||
		strings.TrimSpace(input.Description) == "" ||
		input.HappenedAt.IsZero() {
		return ErrMutationValidation
	}
	if input.Type != "masuk" && input.Type != "keluar" {
		return ErrMutationValidation
	}
	if input.RelatedToDebt {
		return ErrDebtRelationNotSupported
	}
	return nil
}

func sanitizeInput(input UpsertInput) UpsertInput {
	input.WalletID = strings.TrimSpace(input.WalletID)
	input.Amount = strings.TrimSpace(input.Amount)
	input.Description = strings.TrimSpace(input.Description)
	return input
}
