package recurring

import (
	"context"
	"testing"
	"time"
)

type stubRepository struct {
	createFn     func(ctx context.Context, rule *Rule) error
	listFn       func(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	getByIDFn    func(ctx context.Context, userID, ruleID string) (*Rule, error)
	updateFn     func(ctx context.Context, rule *Rule) (*Rule, error)
	inactivateFn func(ctx context.Context, userID, ruleID string) error
	runDueFn     func(ctx context.Context, userID string, now time.Time) (*RunDueResult, error)
}

func (s *stubRepository) Create(ctx context.Context, rule *Rule) error { return s.createFn(ctx, rule) }
func (s *stubRepository) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	return s.listFn(ctx, userID, options)
}
func (s *stubRepository) GetByID(ctx context.Context, userID, ruleID string) (*Rule, error) {
	return s.getByIDFn(ctx, userID, ruleID)
}
func (s *stubRepository) Update(ctx context.Context, rule *Rule) (*Rule, error) {
	return s.updateFn(ctx, rule)
}
func (s *stubRepository) Inactivate(ctx context.Context, userID, ruleID string) error {
	return s.inactivateFn(ctx, userID, ruleID)
}
func (s *stubRepository) RunDue(ctx context.Context, userID string, now time.Time) (*RunDueResult, error) {
	return s.runDueFn(ctx, userID, now)
}

func TestCreateRejectsInvalidIntervalUnit(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.Create(context.Background(), "user-1", CreateInput{
		WalletID:     "wallet-1",
		Type:         "masuk",
		Amount:       "100.00",
		Description:  "salary",
		IntervalUnit: "yearly",
		StartAt:      time.Now(),
	})
	if err != ErrRuleValidation {
		t.Fatalf("expected ErrRuleValidation, got %v", err)
	}
}

func TestRunDueUsesProvidedTime(t *testing.T) {
	expected := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	service := NewService(&stubRepository{
		runDueFn: func(ctx context.Context, userID string, now time.Time) (*RunDueResult, error) {
			if !now.Equal(expected) {
				t.Fatalf("expected %v, got %v", expected, now)
			}
			return &RunDueResult{}, nil
		},
	})

	if _, err := service.RunDue(context.Background(), "user-1", expected); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
