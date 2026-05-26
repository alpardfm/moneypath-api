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
	updateSettingsFn func(ctx context.Context, userID, preferredCurrency, timezone, dateFormat, weekStartDay string) (*User, error)
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
func (s *stubRepository) UpdateSettings(ctx context.Context, userID, preferredCurrency, timezone, dateFormat, weekStartDay string) (*User, error) {
	if s.updateSettingsFn == nil {
		return nil, ErrUserNotFound
	}
	return s.updateSettingsFn(ctx, userID, preferredCurrency, timezone, dateFormat, weekStartDay)
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

func TestRegisterRejectsEmptyEmail(t *testing.T) {
	service := NewService(&stubRepository{}, NewTokenManager("secret"))
	_, err := service.Register(context.Background(), RegisterInput{
		Email:    "",
		Username: "john",
		Password: "password123",
		FullName: "John Doe",
	})
	if err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestRegisterRejectsEmptyUsername(t *testing.T) {
	service := NewService(&stubRepository{}, NewTokenManager("secret"))
	_, err := service.Register(context.Background(), RegisterInput{
		Email:    "john@mail.com",
		Username: "   ",
		Password: "password123",
		FullName: "John Doe",
	})
	if err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestRegisterRejectsEmptyPassword(t *testing.T) {
	service := NewService(&stubRepository{}, NewTokenManager("secret"))
	_, err := service.Register(context.Background(), RegisterInput{
		Email:    "john@mail.com",
		Username: "john",
		Password: "",
		FullName: "John Doe",
	})
	if err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestRegisterRejectsEmptyFullName(t *testing.T) {
	service := NewService(&stubRepository{}, NewTokenManager("secret"))
	_, err := service.Register(context.Background(), RegisterInput{
		Email:    "john@mail.com",
		Username: "john",
		Password: "password123",
		FullName: "   ",
	})
	if err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestRegisterNormalizesEmailAndUsername(t *testing.T) {
	var captured *User
	repo := &stubRepository{
		createFn: func(ctx context.Context, user *User) error {
			user.ID = "user-1"
			captured = user
			return nil
		},
	}
	service := NewService(repo, NewTokenManager("secret"))

	_, err := service.Register(context.Background(), RegisterInput{
		Email:    "  JOHN@MAIL.COM  ",
		Username: "  JohnDoe  ",
		Password: "password123",
		FullName: "  John Doe  ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if captured.Email != "john@mail.com" {
		t.Fatalf("expected lowercase trimmed email, got %q", captured.Email)
	}
	if captured.Username != "johndoe" {
		t.Fatalf("expected lowercase trimmed username, got %q", captured.Username)
	}
	if captured.FullName != "John Doe" {
		t.Fatalf("expected trimmed full name, got %q", captured.FullName)
	}
}

func TestLoginRejectsEmptyCredentials(t *testing.T) {
	service := NewService(&stubRepository{}, NewTokenManager("secret"))
	_, err := service.Login(context.Background(), LoginInput{
		EmailOrUsername: "",
		Password:        "password123",
	})
	if err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestLoginRejectsEmptyPassword(t *testing.T) {
	service := NewService(&stubRepository{}, NewTokenManager("secret"))
	_, err := service.Login(context.Background(), LoginInput{
		EmailOrUsername: "john",
		Password:        "",
	})
	if err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestLoginRejectsNonExistentUser(t *testing.T) {
	repo := &stubRepository{
		getEmailOrUserFn: func(ctx context.Context, value string) (*User, error) {
			return nil, ErrUserNotFound
		},
	}
	service := NewService(repo, NewTokenManager("secret"))

	_, err := service.Login(context.Background(), LoginInput{
		EmailOrUsername: "nobody",
		Password:        "password123",
	})
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginSuccessReturnsToken(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	repo := &stubRepository{
		getEmailOrUserFn: func(ctx context.Context, value string) (*User, error) {
			return &User{ID: "user-1", Email: "john@mail.com", PasswordHash: string(hash)}, nil
		},
	}
	service := NewService(repo, NewTokenManager("secret"))

	result, err := service.Login(context.Background(), LoginInput{
		EmailOrUsername: "john@mail.com",
		Password:        "correct-password",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token to be generated")
	}
	if result.User.ID != "user-1" {
		t.Fatalf("expected user-1, got %q", result.User.ID)
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
