package category

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements category persistence using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a category repository backed by PostgreSQL.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, category *Category) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (user_id, name, category_type)
		VALUES ($1, $2, $3)
		RETURNING id, is_active, deleted_at, created_at, updated_at
	`, category.UserID, category.Name, category.Type).Scan(
		&category.ID,
		&category.IsActive,
		&category.DeletedAt,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err == nil {
		return nil
	}
	return mapConstraintError(err)
}

func (r *PostgresRepository) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	whereParts := []string{"user_id = $1", "deleted_at IS NULL", "is_active = TRUE"}
	args := []any{userID}
	argIndex := 2

	if options.Type != "" {
		whereParts = append(whereParts, "category_type = $2")
		args = append(args, options.Type)
		argIndex++
	}

	var totalItems int
	countQuery := `
		SELECT COUNT(1)
		FROM categories
		WHERE ` + strings.Join(whereParts, " AND ")
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalItems); err != nil {
		return nil, err
	}

	offset := (options.Page - 1) * options.PageSize
	args = append(args, options.PageSize, offset)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, user_id, name, category_type, is_active, deleted_at, created_at, updated_at
		FROM categories
		WHERE %s
		ORDER BY category_type ASC, name ASC, created_at ASC
		LIMIT $%d OFFSET $%d
	`, strings.Join(whereParts, " AND "), argIndex, argIndex+1), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Category
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.Type, &item.IsActive, &item.DeletedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return &ListResult{
		Items:      items,
		TotalItems: totalItems,
	}, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, userID, categoryID string) (*Category, error) {
	item := &Category{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, category_type, is_active, deleted_at, created_at, updated_at
		FROM categories
		WHERE id = $1 AND user_id = $2
	`, categoryID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.Name,
		&item.Type,
		&item.IsActive,
		&item.DeletedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) Update(ctx context.Context, userID, categoryID, name, categoryType string) (*Category, error) {
	item := &Category{}
	err := r.pool.QueryRow(ctx, `
		UPDATE categories
		SET name = $3, category_type = $4, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, category_type, is_active, deleted_at, created_at, updated_at
	`, categoryID, userID, name, categoryType).Scan(
		&item.ID,
		&item.UserID,
		&item.Name,
		&item.Type,
		&item.IsActive,
		&item.DeletedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, mapConstraintError(err)
	}
	return item, nil
}

func (r *PostgresRepository) Inactivate(ctx context.Context, userID, categoryID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE categories
		SET is_active = FALSE, deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, categoryID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func mapConstraintError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "idx_categories_user_id_name_type_active_unique") {
			return ErrCategoryNameAlreadyUsed
		}
	}
	return err
}
