package analytics

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements analytics read queries using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed analytics repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetMonthlyReport(ctx context.Context, userID string, months int) (*MonthlyReport, error) {
	rows, err := r.pool.Query(ctx, `
		WITH month_series AS (
			SELECT generate_series(
				date_trunc('month', CURRENT_DATE) - (($2 - 1) * interval '1 month'),
				date_trunc('month', CURRENT_DATE),
				interval '1 month'
			) AS month_start
		)
		SELECT
			to_char(month_series.month_start, 'YYYY-MM') AS month,
			COALESCE(SUM(CASE WHEN mutations.mutation_type = 'masuk' THEN mutations.amount END), 0)::text AS total_incoming,
			COALESCE(SUM(CASE WHEN mutations.mutation_type = 'keluar' THEN mutations.amount END), 0)::text AS total_outgoing,
			(
				COALESCE(SUM(CASE WHEN mutations.mutation_type = 'masuk' THEN mutations.amount END), 0) -
				COALESCE(SUM(CASE WHEN mutations.mutation_type = 'keluar' THEN mutations.amount END), 0)
			)::text AS net_flow
		FROM month_series
		LEFT JOIN mutations
			ON mutations.user_id = $1
			AND mutations.happened_at >= month_series.month_start
			AND mutations.happened_at < month_series.month_start + interval '1 month'
		GROUP BY month_series.month_start
		ORDER BY month_series.month_start ASC
	`, userID, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &MonthlyReport{Months: months}
	for rows.Next() {
		var item MonthlyPoint
		if err := rows.Scan(&item.Month, &item.TotalIncoming, &item.TotalOutgoing, &item.NetFlow); err != nil {
			return nil, err
		}
		report.Items = append(report.Items, item)
	}

	return report, rows.Err()
}
