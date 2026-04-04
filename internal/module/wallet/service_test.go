package wallet

import (
	"context"
	"testing"
)

type stubRepository struct {
	createFn     func(ctx context.Context, wallet *Wallet) error
	listFn       func(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	getByIDFn    func(ctx context.Context, userID, walletID string) (*Wallet, error)
	updateNameFn func(ctx context.Context, userID, walletID, name string) (*Wallet, error)
	inactiveFn   func(ctx context.Context, userID, walletID string) error
}

func (s *stubRepository) Create(ctx context.Context, wallet *Wallet) error {
	return s.createFn(ctx, wallet)
}
func (s *stubRepository) ListActive(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	return s.listFn(ctx, userID, options)
}
func (s *stubRepository) GetByID(ctx context.Context, userID, walletID string) (*Wallet, error) {
	return s.getByIDFn(ctx, userID, walletID)
}
func (s *stubRepository) UpdateName(ctx context.Context, userID, walletID, name string) (*Wallet, error) {
	return s.updateNameFn(ctx, userID, walletID, name)
}
func (s *stubRepository) Inactivate(ctx context.Context, userID, walletID string) error {
	return s.inactiveFn(ctx, userID, walletID)
}

func TestCreateRejectsEmptyName(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.Create(context.Background(), "user-1", CreateInput{Name: "   "})
	if err != ErrWalletValidation {
		t.Fatalf("expected ErrWalletValidation, got %v", err)
	}
}

func TestInactivateReturnsBalanceNotZero(t *testing.T) {
	service := NewService(&stubRepository{
		inactiveFn: func(ctx context.Context, userID, walletID string) error {
			return ErrWalletBalanceNotZero
		},
	})

	err := service.Inactivate(context.Background(), "user-1", "wallet-1")
	if err != ErrWalletBalanceNotZero {
		t.Fatalf("expected ErrWalletBalanceNotZero, got %v", err)
	}
}

func TestListActiveUsesDefaultPagination(t *testing.T) {
	service := NewService(&stubRepository{
		listFn: func(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
			if options.Page != 1 {
				t.Fatalf("expected default page 1, got %d", options.Page)
			}
			if options.PageSize != 20 {
				t.Fatalf("expected default page size 20, got %d", options.PageSize)
			}
			return &ListResult{}, nil
		},
	})

	if _, err := service.ListActive(context.Background(), "user-1", ListOptions{}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
