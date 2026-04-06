package category

import "context"

// ListOptions contains category list query parameters.
type ListOptions struct {
	Page     int
	PageSize int
	Type     string
}

// ListResult contains paginated category data.
type ListResult struct {
	Items      []Category
	TotalItems int
}

// Repository defines the persistence contract for category flows.
type Repository interface {
	Create(ctx context.Context, category *Category) error
	List(ctx context.Context, userID string, options ListOptions) (*ListResult, error)
	GetByID(ctx context.Context, userID, categoryID string) (*Category, error)
	Update(ctx context.Context, userID, categoryID, name, categoryType string) (*Category, error)
	Inactivate(ctx context.Context, userID, categoryID string) error
}
