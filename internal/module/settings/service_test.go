package settings

import (
	"context"
	"testing"

	"github.com/alpardfm/moneypath-api/internal/module/auth"
)

type stubRepository struct {
	getByIDFn        func(ctx context.Context, userID string) (*auth.User, error)
	updateSettingsFn func(ctx context.Context, userID, preferredCurrency, timezone, dateFormat, weekStartDay string) (*auth.User, error)
}

func (s *stubRepository) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	return s.getByIDFn(ctx, userID)
}
func (s *stubRepository) UpdateSettings(ctx context.Context, userID, preferredCurrency, timezone, dateFormat, weekStartDay string) (*auth.User, error) {
	return s.updateSettingsFn(ctx, userID, preferredCurrency, timezone, dateFormat, weekStartDay)
}

func TestUpdateRejectsInvalidWeekStartDay(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.Update(context.Background(), "user-1", UpdateInput{
		PreferredCurrency: "IDR",
		Timezone:          "Asia/Jakarta",
		DateFormat:        "YYYY-MM-DD",
		WeekStartDay:      "friday",
	})
	if err != auth.ErrValidation {
		t.Fatalf("expected auth.ErrValidation, got %v", err)
	}
}

func TestUpdateNormalizesCurrency(t *testing.T) {
	service := NewService(&stubRepository{
		updateSettingsFn: func(ctx context.Context, userID, preferredCurrency, timezone, dateFormat, weekStartDay string) (*auth.User, error) {
			if preferredCurrency != "USD" {
				t.Fatalf("expected USD, got %q", preferredCurrency)
			}
			return &auth.User{ID: userID, PreferredCurrency: preferredCurrency, Timezone: timezone, DateFormat: dateFormat, WeekStartDay: weekStartDay}, nil
		},
	})

	_, err := service.Update(context.Background(), "user-1", UpdateInput{
		PreferredCurrency: "usd",
		Timezone:          "Asia/Jakarta",
		DateFormat:        "YYYY-MM-DD",
		WeekStartDay:      "monday",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
