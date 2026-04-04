package debt

import (
	"context"
	"strings"
)

// Service contains debt use cases.
type Service struct {
	repo Repository
}

// NewService creates a debt service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new debt.
func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*Debt, error) {
	name := strings.TrimSpace(input.Name)
	principal := strings.TrimSpace(input.Principal)
	if name == "" || principal == "" {
		return nil, ErrDebtValidation
	}

	item := &Debt{
		UserID:          userID,
		Name:            name,
		PrincipalAmount: principal,
		RemainingAmount: principal,
		TenorValue:      input.TenorValue,
		TenorUnit:       normalizeStringPointer(input.TenorUnit),
		PaymentAmount:   normalizeStringPointer(input.PaymentAmount),
		Note:            normalizeStringPointer(input.Note),
		Status:          "active",
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	item.Status = deriveStatus(item.RemainingAmount, item.IsActive)
	return item, nil
}

// List lists debts for the authenticated user.
func (s *Service) List(ctx context.Context, userID string) ([]Debt, error) {
	items, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Status = deriveStatus(items[i].RemainingAmount, items[i].IsActive)
	}
	return items, nil
}

// GetByID returns a debt by id for the authenticated user.
func (s *Service) GetByID(ctx context.Context, userID, debtID string) (*Debt, error) {
	item, err := s.repo.GetByID(ctx, userID, debtID)
	if err != nil {
		return nil, err
	}
	item.Status = deriveStatus(item.RemainingAmount, item.IsActive)
	return item, nil
}

// Update updates editable debt metadata.
func (s *Service) Update(ctx context.Context, userID, debtID string, input UpdateInput) (*Debt, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrDebtValidation
	}

	item := &Debt{
		ID:            debtID,
		UserID:        userID,
		Name:          name,
		TenorValue:    input.TenorValue,
		TenorUnit:     normalizeStringPointer(input.TenorUnit),
		PaymentAmount: normalizeStringPointer(input.PaymentAmount),
		Note:          normalizeStringPointer(input.Note),
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return nil, err
	}
	updated.Status = deriveStatus(updated.RemainingAmount, updated.IsActive)
	return updated, nil
}

// Inactivate inactivates a paid debt.
func (s *Service) Inactivate(ctx context.Context, userID, debtID string) error {
	return s.repo.Inactivate(ctx, userID, debtID)
}

func deriveStatus(remaining string, isActive bool) string {
	if !isActive {
		return "inactive"
	}
	if remaining == "0.00" || remaining == "0" {
		return "lunas"
	}
	return "active"
}

func normalizeStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
