package debt

import (
	"context"
	"testing"
)

type stubRepository struct {
	createFn       func(ctx context.Context, debt *Debt) error
	listFn         func(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	listArchivedFn func(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	getByIDFn      func(ctx context.Context, userID, debtID string) (*Debt, error)
	updateFn       func(ctx context.Context, debt *Debt) (*Debt, error)
	inactiveFn     func(ctx context.Context, userID, debtID string) error
}

func (s *stubRepository) Create(ctx context.Context, debt *Debt) error { return s.createFn(ctx, debt) }
func (s *stubRepository) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	return s.listFn(ctx, userID, options)
}
func (s *stubRepository) ListArchived(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	if s.listArchivedFn == nil {
		return &ListResult{}, nil
	}
	return s.listArchivedFn(ctx, userID, options)
}
func (s *stubRepository) GetByID(ctx context.Context, userID, debtID string) (*Debt, error) {
	return s.getByIDFn(ctx, userID, debtID)
}
func (s *stubRepository) Update(ctx context.Context, debt *Debt) (*Debt, error) {
	return s.updateFn(ctx, debt)
}
func (s *stubRepository) Inactivate(ctx context.Context, userID, debtID string) error {
	return s.inactiveFn(ctx, userID, debtID)
}

func TestCreateRejectsEmptyFields(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", CreateInput{Name: " ", Principal: ""})
	if err != ErrDebtValidation {
		t.Fatalf("expected ErrDebtValidation, got %v", err)
	}
}

func TestDeriveStatusLunas(t *testing.T) {
	if got := deriveStatus("0.00", true); got != "lunas" {
		t.Fatalf("expected lunas, got %q", got)
	}
}

func TestInactivateReturnsRemainingNotZero(t *testing.T) {
	service := NewService(&stubRepository{
		inactiveFn: func(ctx context.Context, userID, debtID string) error {
			return ErrDebtRemainingNotZero
		},
	})
	if err := service.Inactivate(context.Background(), "user-1", "debt-1"); err != ErrDebtRemainingNotZero {
		t.Fatalf("expected ErrDebtRemainingNotZero, got %v", err)
	}
}

func TestListUsesDefaultPagination(t *testing.T) {
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

	if _, err := service.List(context.Background(), "user-1", ListOptions{}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
