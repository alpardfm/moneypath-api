package summary

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements summary read queries using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed summary repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetReport(ctx context.Context, userID string, filter Filter) (*Report, error) {
	report := &Report{From: filter.From, To: filter.To}
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT SUM(balance)::text FROM wallets WHERE user_id = $1 AND is_active = TRUE AND deleted_at IS NULL), '0'),
			COALESCE((SELECT SUM(remaining_amount)::text FROM debts WHERE user_id = $1 AND is_active = TRUE AND deleted_at IS NULL), '0'),
			COALESCE((
				SELECT SUM(amount)::text
				FROM mutations
				WHERE user_id = $1 AND mutation_type = 'masuk'
				  AND ($2::timestamptz IS NULL OR happened_at >= $2)
				  AND ($3::timestamptz IS NULL OR happened_at <= $3)
			), '0'),
			COALESCE((
				SELECT SUM(amount)::text
				FROM mutations
				WHERE user_id = $1 AND mutation_type = 'keluar'
				  AND ($2::timestamptz IS NULL OR happened_at >= $2)
				  AND ($3::timestamptz IS NULL OR happened_at <= $3)
			), '0'),
			COALESCE((
				SELECT (COALESCE(SUM(CASE WHEN mutation_type = 'masuk' THEN amount END), 0) -
				        COALESCE(SUM(CASE WHEN mutation_type = 'keluar' THEN amount END), 0))::text
				FROM mutations
				WHERE user_id = $1
				  AND ($2::timestamptz IS NULL OR happened_at >= $2)
				  AND ($3::timestamptz IS NULL OR happened_at <= $3)
			), '0')
	`, userID, filter.From, filter.To).Scan(
		&report.TotalAssets,
		&report.TotalDebts,
		&report.TotalIncoming,
		&report.TotalOutgoing,
		&report.NetFlow,
	)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, balance::text
		FROM wallets
		WHERE user_id = $1 AND is_active = TRUE AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item WalletBalance
		if err := rows.Scan(&item.WalletID, &item.Name, &item.Balance); err != nil {
			return nil, err
		}
		report.Wallets = append(report.Wallets, item)
	}
	return report, rows.Err()
}
