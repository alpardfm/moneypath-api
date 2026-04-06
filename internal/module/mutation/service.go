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
func (s *Service) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.PageSize <= 0 {
		options.PageSize = 20
	}
	if options.SortBy == "" {
		options.SortBy = "happened_at"
	}
	if options.SortBy != "happened_at" && options.SortBy != "created_at" && options.SortBy != "amount" {
		return nil, ErrMutationValidation
	}
	if options.SortDirection == "" {
		options.SortDirection = "desc"
	}
	if options.SortDirection != "asc" && options.SortDirection != "desc" {
		return nil, ErrMutationValidation
	}
	if options.Type != "" && options.Type != "masuk" && options.Type != "keluar" {
		return nil, ErrMutationValidation
	}
	return s.repo.List(ctx, userID, options)
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

	if !input.RelatedToDebt {
		if input.DebtID != nil || input.NewDebt != nil {
			return ErrInvalidDebtRelation
		}
		return nil
	}

	if input.Type == "keluar" {
		if input.DebtID == nil || strings.TrimSpace(*input.DebtID) == "" || input.NewDebt != nil {
			return ErrInvalidDebtRelation
		}
		return nil
	}

	if input.Type == "masuk" {
		hasDebtID := input.DebtID != nil && strings.TrimSpace(*input.DebtID) != ""
		hasNewDebt := input.NewDebt != nil
		if hasDebtID == hasNewDebt {
			return ErrInvalidDebtRelation
		}
		if hasNewDebt {
			if strings.TrimSpace(input.NewDebt.Name) == "" || strings.TrimSpace(input.NewDebt.Principal) == "" {
				return ErrMutationValidation
			}
		}
		return nil
	}

	return nil
}

func sanitizeInput(input UpsertInput) UpsertInput {
	input.WalletID = strings.TrimSpace(input.WalletID)
	input.Amount = strings.TrimSpace(input.Amount)
	input.Description = strings.TrimSpace(input.Description)
	if input.CategoryID != nil {
		categoryID := strings.TrimSpace(*input.CategoryID)
		if categoryID == "" {
			input.CategoryID = nil
		} else {
			input.CategoryID = &categoryID
		}
	}
	if input.DebtID != nil {
		debtID := strings.TrimSpace(*input.DebtID)
		input.DebtID = &debtID
	}
	if input.NewDebt != nil {
		input.NewDebt.Name = strings.TrimSpace(input.NewDebt.Name)
		input.NewDebt.Principal = strings.TrimSpace(input.NewDebt.Principal)
		if input.NewDebt.TenorUnit != nil {
			trimmed := strings.TrimSpace(*input.NewDebt.TenorUnit)
			input.NewDebt.TenorUnit = &trimmed
		}
		if input.NewDebt.PaymentAmount != nil {
			trimmed := strings.TrimSpace(*input.NewDebt.PaymentAmount)
			input.NewDebt.PaymentAmount = &trimmed
		}
		if input.NewDebt.Note != nil {
			trimmed := strings.TrimSpace(*input.NewDebt.Note)
			input.NewDebt.Note = &trimmed
		}
	}
	return input
}
