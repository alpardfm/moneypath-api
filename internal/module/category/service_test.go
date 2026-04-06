package category

import (
	"context"
	"testing"
)

type stubRepository struct {
	createFn     func(ctx context.Context, category *Category) error
	listFn       func(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	getByIDFn    func(ctx context.Context, userID, categoryID string) (*Category, error)
	updateFn     func(ctx context.Context, userID, categoryID, name, categoryType string) (*Category, error)
	inactivateFn func(ctx context.Context, userID, categoryID string) error
}

func (s *stubRepository) Create(ctx context.Context, category *Category) error {
	return s.createFn(ctx, category)
}
func (s *stubRepository) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	return s.listFn(ctx, userID, options)
}
func (s *stubRepository) GetByID(ctx context.Context, userID, categoryID string) (*Category, error) {
	return s.getByIDFn(ctx, userID, categoryID)
}
func (s *stubRepository) Update(ctx context.Context, userID, categoryID, name, categoryType string) (*Category, error) {
	return s.updateFn(ctx, userID, categoryID, name, categoryType)
}
func (s *stubRepository) Inactivate(ctx context.Context, userID, categoryID string) error {
	return s.inactivateFn(ctx, userID, categoryID)
}

func TestCreateRejectsInvalidType(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.Create(context.Background(), "user-1", CreateInput{Name: "Salary", Type: "other"})
	if err != ErrCategoryValidation {
		t.Fatalf("expected ErrCategoryValidation, got %v", err)
	}
}

func TestListUsesDefaultPagination(t *testing.T) {
	service := NewService(&stubRepository{
		listFn: func(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
			if options.Page != 1 || options.PageSize != 20 {
				t.Fatalf("unexpected pagination: %+v", options)
			}
			return &ListResult{}, nil
		},
	})

	if _, err := service.List(context.Background(), "user-1", ListOptions{}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
