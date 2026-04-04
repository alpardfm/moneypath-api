package wallet

import (
	"context"
	"strings"
)

// Service contains wallet use cases.
type Service struct {
	repo Repository
}

// NewService creates a wallet service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new wallet for the authenticated user.
func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*Wallet, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrWalletValidation
	}

	wallet := &Wallet{
		UserID: userID,
		Name:   name,
	}
	if err := s.repo.Create(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

// ListActive returns active wallets for the authenticated user.
func (s *Service) ListActive(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.PageSize <= 0 {
		options.PageSize = 20
	}
	return s.repo.ListActive(ctx, userID, options)
}

// GetByID returns a wallet owned by the authenticated user.
func (s *Service) GetByID(ctx context.Context, userID, walletID string) (*Wallet, error) {
	return s.repo.GetByID(ctx, userID, walletID)
}

// Update updates wallet metadata.
func (s *Service) Update(ctx context.Context, userID, walletID string, input UpdateInput) (*Wallet, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrWalletValidation
	}
	return s.repo.UpdateName(ctx, userID, walletID, name)
}

// Inactivate hides a wallet from active selection.
func (s *Service) Inactivate(ctx context.Context, userID, walletID string) error {
	return s.repo.Inactivate(ctx, userID, walletID)
}
