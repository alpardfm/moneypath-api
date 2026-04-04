package auth

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Service contains auth use cases.
type Service struct {
	repo   Repository
	tokens *TokenManager
}

// NewService creates an auth service.
func NewService(repo Repository, tokens *TokenManager) *Service {
	return &Service{repo: repo, tokens: tokens}
}

// Register creates a new user account and returns a bearer token.
func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	input.Email = normalize(input.Email)
	input.Username = normalize(input.Username)
	input.FullName = strings.TrimSpace(input.FullName)

	if input.Email == "" || input.Username == "" || input.FullName == "" || input.Password == "" {
		return nil, ErrValidation
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hash),
		FullName:     input.FullName,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

// Login authenticates a user and returns a bearer token.
func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	if normalize(input.EmailOrUsername) == "" || input.Password == "" {
		return nil, ErrValidation
	}

	user, err := s.repo.GetUserByEmailOrUsername(ctx, normalize(input.EmailOrUsername))
	if err != nil {
		if err == ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

func normalize(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}
