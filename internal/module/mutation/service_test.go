package mutation

import (
	"context"
	"testing"
	"time"
)

type stubRepository struct {
	createFn func(ctx context.Context, userID string, input UpsertInput) (*Mutation, error)
	listFn   func(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	getFn    func(ctx context.Context, userID, mutationID string) (*Mutation, error)
	updateFn func(ctx context.Context, userID, mutationID string, input UpsertInput) (*Mutation, error)
	deleteFn func(ctx context.Context, userID, mutationID string) error
}

func (s *stubRepository) Create(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
	return s.createFn(ctx, userID, input)
}
func (s *stubRepository) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	return s.listFn(ctx, userID, options)
}
func (s *stubRepository) GetByID(ctx context.Context, userID, mutationID string) (*Mutation, error) {
	return s.getFn(ctx, userID, mutationID)
}
func (s *stubRepository) Update(ctx context.Context, userID, mutationID string, input UpsertInput) (*Mutation, error) {
	return s.updateFn(ctx, userID, mutationID, input)
}
func (s *stubRepository) Delete(ctx context.Context, userID, mutationID string) error {
	return s.deleteFn(ctx, userID, mutationID)
}

func TestCreateRejectsDebtRelationInPhaseFive(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:      "wallet-1",
		Type:          "masuk",
		Amount:        "10.00",
		Description:   "salary",
		RelatedToDebt: true,
		DebtID:        nil,
		HappenedAt:    time.Now(),
	})
	if err != ErrInvalidDebtRelation {
		t.Fatalf("expected ErrInvalidDebtRelation, got %v", err)
	}
}

func TestCreateRejectsEmptyWalletID(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:    "  ",
		Type:        "masuk",
		Amount:      "10.00",
		Description: "salary",
		HappenedAt:  time.Now(),
	})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestCreateRejectsEmptyAmount(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:    "wallet-1",
		Type:        "masuk",
		Amount:      "",
		Description: "salary",
		HappenedAt:  time.Now(),
	})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestCreateRejectsEmptyDescription(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:    "wallet-1",
		Type:        "masuk",
		Amount:      "10.00",
		Description: "   ",
		HappenedAt:  time.Now(),
	})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestCreateRejectsZeroHappenedAt(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:    "wallet-1",
		Type:        "masuk",
		Amount:      "10.00",
		Description: "salary",
	})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestCreateRejectsInvalidType(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:    "wallet-1",
		Type:        "invalid",
		Amount:      "10.00",
		Description: "salary",
		HappenedAt:  time.Now(),
	})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestCreateRejectsDebtFieldsWhenNotRelated(t *testing.T) {
	debtID := "debt-1"
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:      "wallet-1",
		Type:          "masuk",
		Amount:        "10.00",
		Description:   "salary",
		RelatedToDebt: false,
		DebtID:        &debtID,
		HappenedAt:    time.Now(),
	})
	if err != ErrInvalidDebtRelation {
		t.Fatalf("expected ErrInvalidDebtRelation, got %v", err)
	}
}

func TestCreateRejectsKeluarDebtWithoutDebtID(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:      "wallet-1",
		Type:          "keluar",
		Amount:        "50.00",
		Description:   "debt payment",
		RelatedToDebt: true,
		DebtID:        nil,
		HappenedAt:    time.Now(),
	})
	if err != ErrInvalidDebtRelation {
		t.Fatalf("expected ErrInvalidDebtRelation, got %v", err)
	}
}

func TestCreateAcceptsKeluarDebtWithDebtID(t *testing.T) {
	debtID := "debt-1"
	called := false
	service := NewService(&stubRepository{
		createFn: func(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
			called = true
			return &Mutation{ID: "m-1", UserID: userID}, nil
		},
	})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:      "wallet-1",
		Type:          "keluar",
		Amount:        "50.00",
		Description:   "debt payment",
		RelatedToDebt: true,
		DebtID:        &debtID,
		HappenedAt:    time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected repository create to be called")
	}
}

func TestCreateRejectsMasukDebtWithBothDebtIDAndNewDebt(t *testing.T) {
	debtID := "debt-1"
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:      "wallet-1",
		Type:          "masuk",
		Amount:        "100.00",
		Description:   "loan",
		RelatedToDebt: true,
		DebtID:        &debtID,
		NewDebt:       &NewDebtInput{Name: "Laptop", Principal: "120.00"},
		HappenedAt:    time.Now(),
	})
	if err != ErrInvalidDebtRelation {
		t.Fatalf("expected ErrInvalidDebtRelation, got %v", err)
	}
}

func TestCreateRejectsMasukDebtWithNeitherDebtIDNorNewDebt(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:      "wallet-1",
		Type:          "masuk",
		Amount:        "100.00",
		Description:   "loan",
		RelatedToDebt: true,
		DebtID:        nil,
		NewDebt:       nil,
		HappenedAt:    time.Now(),
	})
	if err != ErrInvalidDebtRelation {
		t.Fatalf("expected ErrInvalidDebtRelation, got %v", err)
	}
}

func TestCreateRejectsNewDebtWithEmptyName(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:      "wallet-1",
		Type:          "masuk",
		Amount:        "100.00",
		Description:   "loan",
		RelatedToDebt: true,
		NewDebt:       &NewDebtInput{Name: "  ", Principal: "120.00"},
		HappenedAt:    time.Now(),
	})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestCreateSanitizesInput(t *testing.T) {
	var captured UpsertInput
	service := NewService(&stubRepository{
		createFn: func(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
			captured = input
			return &Mutation{ID: "m-1"}, nil
		},
	})
	categoryID := "  cat-1  "
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:    "  wallet-1  ",
		Type:        "masuk",
		Amount:      "  10.00  ",
		Description: "  salary  ",
		CategoryID:  &categoryID,
		HappenedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if captured.WalletID != "wallet-1" {
		t.Fatalf("expected trimmed wallet_id, got %q", captured.WalletID)
	}
	if captured.Amount != "10.00" {
		t.Fatalf("expected trimmed amount, got %q", captured.Amount)
	}
	if captured.Description != "salary" {
		t.Fatalf("expected trimmed description, got %q", captured.Description)
	}
	if *captured.CategoryID != "cat-1" {
		t.Fatalf("expected trimmed category_id, got %q", *captured.CategoryID)
	}
}

func TestCreateNilsCategoryIDWhenEmpty(t *testing.T) {
	var captured UpsertInput
	service := NewService(&stubRepository{
		createFn: func(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
			captured = input
			return &Mutation{ID: "m-1"}, nil
		},
	})
	emptyCategory := "   "
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:    "wallet-1",
		Type:        "keluar",
		Amount:      "5.00",
		Description: "snack",
		CategoryID:  &emptyCategory,
		HappenedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if captured.CategoryID != nil {
		t.Fatalf("expected nil category_id after sanitize, got %v", captured.CategoryID)
	}
}

func TestCreateAcceptsIncomingBorrowNew(t *testing.T) {
	called := false
	service := NewService(&stubRepository{
		createFn: func(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
			called = true
			return &Mutation{ID: "mutation-1", UserID: userID}, nil
		},
	})
	_, err := service.Create(context.Background(), "user-1", UpsertInput{
		WalletID:      "wallet-1",
		Type:          "masuk",
		Amount:        "100.00",
		Description:   "loan disbursement",
		RelatedToDebt: true,
		NewDebt: &NewDebtInput{
			Name:      "Laptop",
			Principal: "120.00",
		},
		HappenedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected repository create to be called")
	}
}

func TestDeleteAlwaysRejected(t *testing.T) {
	service := NewService(&stubRepository{
		deleteFn: func(ctx context.Context, userID, mutationID string) error {
			return ErrMutationDeleteNotAllowed
		},
	})
	if err := service.Delete(context.Background(), "user-1", "mutation-1"); err != ErrMutationDeleteNotAllowed {
		t.Fatalf("expected ErrMutationDeleteNotAllowed, got %v", err)
	}
}

func TestListRejectsInvalidSort(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.List(context.Background(), "user-1", ListOptions{SortBy: "unknown"})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestListUsesDefaultPaginationAndSort(t *testing.T) {
	service := NewService(&stubRepository{
		listFn: func(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
			if options.Page != 1 || options.PageSize != 20 {
				t.Fatalf("unexpected pagination: %+v", options)
			}
			if options.SortBy != "happened_at" || options.SortDirection != "desc" {
				t.Fatalf("unexpected sort defaults: %+v", options)
			}
			return &ListResult{}, nil
		},
	})

	if _, err := service.List(context.Background(), "user-1", ListOptions{}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListRejectsInvalidSortDirection(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.List(context.Background(), "user-1", ListOptions{SortDirection: "random"})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestListRejectsInvalidType(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.List(context.Background(), "user-1", ListOptions{Type: "transfer"})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestListAcceptsValidSortByAmount(t *testing.T) {
	service := NewService(&stubRepository{
		listFn: func(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
			if options.SortBy != "amount" {
				t.Fatalf("expected sort by amount, got %q", options.SortBy)
			}
			return &ListResult{}, nil
		},
	})
	_, err := service.List(context.Background(), "user-1", ListOptions{SortBy: "amount"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListAcceptsTypeMasuk(t *testing.T) {
	service := NewService(&stubRepository{
		listFn: func(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
			if options.Type != "masuk" {
				t.Fatalf("expected type masuk, got %q", options.Type)
			}
			return &ListResult{}, nil
		},
	})
	_, err := service.List(context.Background(), "user-1", ListOptions{Type: "masuk"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateValidatesInput(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.Update(context.Background(), "user-1", "m-1", UpsertInput{
		WalletID:    "",
		Type:        "masuk",
		Amount:      "10.00",
		Description: "test",
		HappenedAt:  time.Now(),
	})
	if err != ErrMutationValidation {
		t.Fatalf("expected ErrMutationValidation, got %v", err)
	}
}

func TestUpdateCallsRepoOnValidInput(t *testing.T) {
	called := false
	service := NewService(&stubRepository{
		updateFn: func(ctx context.Context, userID, mutationID string, input UpsertInput) (*Mutation, error) {
			called = true
			if mutationID != "m-1" {
				t.Fatalf("expected mutation id m-1, got %q", mutationID)
			}
			return &Mutation{ID: mutationID}, nil
		},
	})
	_, err := service.Update(context.Background(), "user-1", "m-1", UpsertInput{
		WalletID:    "wallet-1",
		Type:        "keluar",
		Amount:      "25.00",
		Description: "groceries",
		HappenedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected repository update to be called")
	}
}

func TestGetByIDDelegatesToRepo(t *testing.T) {
	service := NewService(&stubRepository{
		getFn: func(ctx context.Context, userID, mutationID string) (*Mutation, error) {
			return &Mutation{ID: mutationID, UserID: userID}, nil
		},
	})
	m, err := service.GetByID(context.Background(), "user-1", "m-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.ID != "m-1" {
		t.Fatalf("expected mutation id m-1, got %q", m.ID)
	}
}
