package notification

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements notification queries using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed notification repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// ListDueRecurring returns active recurring rules due up to the provided time.
func (r *PostgresRepository) ListDueRecurring(ctx context.Context, userID string, until time.Time) ([]RecurringAlert, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, description, amount::text, mutation_type, next_run_at
		FROM recurring_rules
		WHERE user_id = $1
			AND is_active = TRUE
			AND deleted_at IS NULL
			AND next_run_at <= $2
		ORDER BY next_run_at ASC
		LIMIT 10
	`, userID, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RecurringAlert
	for rows.Next() {
		var item RecurringAlert
		if err := rows.Scan(&item.RuleID, &item.Description, &item.Amount, &item.Type, &item.NextRunAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListActiveDebtAlerts returns active debts that still have remaining amount.
func (r *PostgresRepository) ListActiveDebtAlerts(ctx context.Context, userID string, limit int) ([]DebtAlert, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, remaining_amount::text, payment_amount::text, updated_at
		FROM debts
		WHERE user_id = $1
			AND is_active = TRUE
			AND deleted_at IS NULL
			AND remaining_amount > 0
		ORDER BY remaining_amount DESC, updated_at ASC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DebtAlert
	for rows.Next() {
		var item DebtAlert
		if err := rows.Scan(&item.DebtID, &item.Name, &item.RemainingAmount, &item.PaymentAmount, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
