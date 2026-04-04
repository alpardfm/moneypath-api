package auth

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type stubRepository struct {
	createFn         func(ctx context.Context, user *User) error
	getEmailOrUserFn func(ctx context.Context, value string) (*User, error)
	getByIDFn        func(ctx context.Context, userID string) (*User, error)
	updateProfileFn  func(ctx context.Context, userID, email, username, fullName string) (*User, error)
	updatePasswordFn func(ctx context.Context, userID, passwordHash string) error
}

func (s *stubRepository) CreateUser(ctx context.Context, user *User) error {
	return s.createFn(ctx, user)
}
func (s *stubRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return nil, ErrUserNotFound
}
func (s *stubRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return nil, ErrUserNotFound
}
func (s *stubRepository) GetUserByEmailOrUsername(ctx context.Context, value string) (*User, error) {
	return s.getEmailOrUserFn(ctx, value)
}
func (s *stubRepository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	return s.getByIDFn(ctx, userID)
}
func (s *stubRepository) UpdateProfile(ctx context.Context, userID, email, username, fullName string) (*User, error) {
	return s.updateProfileFn(ctx, userID, email, username, fullName)
}
func (s *stubRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return s.updatePasswordFn(ctx, userID, passwordHash)
}

func TestRegisterHashesPasswordAndReturnsToken(t *testing.T) {
	repo := &stubRepository{
		createFn: func(ctx context.Context, user *User) error {
			user.ID = "user-1"
			return nil
		},
	}
	service := NewService(repo, NewTokenManager("secret"))

	result, err := service.Register(context.Background(), RegisterInput{
		Email:    "User@Mail.com",
		Username: "John",
		Password: "password123",
		FullName: "John Doe",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token to be generated")
	}
	if result.User.Email != "user@mail.com" {
		t.Fatalf("expected normalized email, got %q", result.User.Email)
	}
	if result.User.Username != "john" {
		t.Fatalf("expected normalized username, got %q", result.User.Username)
	}
	if result.User.PasswordHash == "password123" {
		t.Fatal("expected password to be hashed")
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	repo := &stubRepository{
		getEmailOrUserFn: func(ctx context.Context, value string) (*User, error) {
			return &User{ID: "user-1", PasswordHash: string(hash)}, nil
		},
	}
	service := NewService(repo, NewTokenManager("secret"))

	_, err := service.Login(context.Background(), LoginInput{
		EmailOrUsername: "john",
		Password:        "wrong-password",
	})
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
