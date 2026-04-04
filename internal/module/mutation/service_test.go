package mutation

import (
	"context"
	"testing"
	"time"
)

type stubRepository struct {
	createFn func(ctx context.Context, userID string, input UpsertInput) (*Mutation, error)
	listFn   func(ctx context.Context, userID string) ([]Mutation, error)
	getFn    func(ctx context.Context, userID, mutationID string) (*Mutation, error)
	updateFn func(ctx context.Context, userID, mutationID string, input UpsertInput) (*Mutation, error)
	deleteFn func(ctx context.Context, userID, mutationID string) error
}

func (s *stubRepository) Create(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
	return s.createFn(ctx, userID, input)
}
func (s *stubRepository) List(ctx context.Context, userID string) ([]Mutation, error) {
	return s.listFn(ctx, userID)
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
