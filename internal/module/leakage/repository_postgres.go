package leakage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements leakage detection queries using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed leakage repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// GetTotalOutgoing returns the total outgoing amount for the selected period.
func (r *PostgresRepository) GetTotalOutgoing(ctx context.Context, userID string, days int) (string, error) {
	var total string
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount)::text, '0')
		FROM mutations
		WHERE user_id = $1
			AND mutation_type = 'keluar'
			AND happened_at >= NOW() - ($2::int * interval '1 day')
	`, userID, days).Scan(&total)
	if err != nil {
		return "", err
	}
	return total, nil
}

// ListCategorySpends returns outgoing aggregates grouped by category for the selected period.
func (r *PostgresRepository) ListCategorySpends(ctx context.Context, userID string, days int) ([]CategorySpend, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			COALESCE(c.id, '') AS category_id,
			COALESCE(c.name, 'Uncategorized') AS category_name,
			COALESCE(SUM(m.amount)::text, '0') AS total_amount,
			COUNT(*)::int AS transaction_count
		FROM mutations m
		LEFT JOIN categories c ON c.id = m.category_id
		WHERE m.user_id = $1
			AND m.mutation_type = 'keluar'
			AND m.happened_at >= NOW() - ($2::int * interval '1 day')
		GROUP BY c.id, c.name
		ORDER BY SUM(m.amount) DESC, COUNT(*) DESC, category_name ASC
		LIMIT 10
	`, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CategorySpend
	for rows.Next() {
		var item CategorySpend
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.TotalAmount, &item.TransactionCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListRepeatedPatterns returns repeated outgoing description patterns for the selected period.
func (r *PostgresRepository) ListRepeatedPatterns(ctx context.Context, userID string, days int) ([]RepeatedPattern, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			MIN(description) AS description,
			COALESCE(SUM(amount)::text, '0') AS total_amount,
			COALESCE(AVG(amount)::text, '0') AS average_amount,
			COUNT(*)::int AS transaction_count
		FROM mutations
		WHERE user_id = $1
			AND mutation_type = 'keluar'
			AND happened_at >= NOW() - ($2::int * interval '1 day')
			AND TRIM(COALESCE(description, '')) <> ''
			AND related_to_debt = FALSE
		GROUP BY LOWER(TRIM(description))
		HAVING COUNT(*) >= 3
		ORDER BY COUNT(*) DESC, SUM(amount) DESC, MIN(description) ASC
		LIMIT 10
	`, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RepeatedPattern
	for rows.Next() {
		var item RepeatedPattern
		if err := rows.Scan(&item.Description, &item.TotalAmount, &item.AverageAmount, &item.TransactionCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
