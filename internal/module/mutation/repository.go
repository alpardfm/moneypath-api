package mutation

import "context"

// Repository defines the persistence contract for mutation flows.
type Repository interface {
	Create(ctx context.Context, userID string, input UpsertInput) (*Mutation, error)
	List(ctx context.Context, userID string) ([]Mutation, error)
	GetByID(ctx context.Context, userID, mutationID string) (*Mutation, error)
	Update(ctx context.Context, userID, mutationID string, input UpsertInput) (*Mutation, error)
	Delete(ctx context.Context, userID, mutationID string) error
}
