package profile

import (
	"context"
	"testing"

	"github.com/alpardfm/moneypath-api/internal/module/auth"
	"golang.org/x/crypto/bcrypt"
)

type stubRepository struct {
	getByIDFn        func(ctx context.Context, userID string) (*auth.User, error)
	updateProfileFn  func(ctx context.Context, userID, email, username, fullName string) (*auth.User, error)
	updatePasswordFn func(ctx context.Context, userID, passwordHash string) error
}

func (s *stubRepository) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	return s.getByIDFn(ctx, userID)
}
func (s *stubRepository) UpdateProfile(ctx context.Context, userID, email, username, fullName string) (*auth.User, error) {
	return s.updateProfileFn(ctx, userID, email, username, fullName)
}
func (s *stubRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return s.updatePasswordFn(ctx, userID, passwordHash)
}

func TestChangePasswordRejectsWrongCurrentPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("current-password"), bcrypt.DefaultCost)
	repo := &stubRepository{
		getByIDFn: func(ctx context.Context, userID string) (*auth.User, error) {
			return &auth.User{ID: userID, PasswordHash: string(hash)}, nil
		},
		updatePasswordFn: func(ctx context.Context, userID, passwordHash string) error {
			t.Fatal("expected update password not to be called")
			return nil
		},
	}

	service := NewService(repo)
	err := service.ChangePassword(context.Background(), "user-1", ChangePasswordInput{
		CurrentPassword: "wrong-password",
		NewPassword:     "new-password",
	})
	if err != auth.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
