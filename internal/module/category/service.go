package category

import (
	"context"
	"strings"
)

// Service contains category use cases.
type Service struct {
	repo Repository
}

// NewService creates a category service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new category for the authenticated user.
func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*Category, error) {
	name := strings.TrimSpace(input.Name)
	categoryType := strings.TrimSpace(input.Type)
	if name == "" || !isValidType(categoryType) {
		return nil, ErrCategoryValidation
	}

	item := &Category{
		UserID: userID,
		Name:   name,
		Type:   categoryType,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// List returns active categories for the authenticated user.
func (s *Service) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.PageSize <= 0 {
		options.PageSize = 20
	}
	options.Type = strings.TrimSpace(options.Type)
	if options.Type != "" && !isValidType(options.Type) {
		return nil, ErrCategoryValidation
	}
	return s.repo.List(ctx, userID, options)
}

// GetByID returns a category owned by the authenticated user.
func (s *Service) GetByID(ctx context.Context, userID, categoryID string) (*Category, error) {
	return s.repo.GetByID(ctx, userID, categoryID)
}

// Update updates category metadata.
func (s *Service) Update(ctx context.Context, userID, categoryID string, input UpdateInput) (*Category, error) {
	name := strings.TrimSpace(input.Name)
	categoryType := strings.TrimSpace(input.Type)
	if name == "" || !isValidType(categoryType) {
		return nil, ErrCategoryValidation
	}
	return s.repo.Update(ctx, userID, categoryID, name, categoryType)
}

// Inactivate hides a category from active selection.
func (s *Service) Inactivate(ctx context.Context, userID, categoryID string) error {
	return s.repo.Inactivate(ctx, userID, categoryID)
}

func isValidType(value string) bool {
	return value == "masuk" || value == "keluar"
}
