package dashboard

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements dashboard read queries using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed dashboard repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetOverview(ctx context.Context, userID string) (*Overview, error) {
	overview := &Overview{}
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT SUM(balance)::text FROM wallets WHERE user_id = $1 AND is_active = TRUE AND deleted_at IS NULL), '0'),
			COALESCE((SELECT SUM(remaining_amount)::text FROM debts WHERE user_id = $1 AND is_active = TRUE AND deleted_at IS NULL), '0'),
			COALESCE((SELECT SUM(amount)::text FROM mutations WHERE user_id = $1 AND mutation_type = 'masuk'), '0'),
			COALESCE((SELECT SUM(amount)::text FROM mutations WHERE user_id = $1 AND mutation_type = 'keluar'), '0'),
			COALESCE((
				SELECT (COALESCE(SUM(CASE WHEN mutation_type = 'masuk' THEN amount END), 0) -
				        COALESCE(SUM(CASE WHEN mutation_type = 'keluar' THEN amount END), 0))::text
				FROM mutations WHERE user_id = $1
			), '0')
	`, userID).Scan(
		&overview.TotalAssets,
		&overview.TotalDebts,
		&overview.TotalIncoming,
		&overview.TotalOutgoing,
		&overview.NetFlow,
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
		overview.Wallets = append(overview.Wallets, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	trendRows, err := r.pool.Query(ctx, `
		WITH month_series AS (
			SELECT generate_series(
				date_trunc('month', CURRENT_DATE) - interval '5 months',
				date_trunc('month', CURRENT_DATE),
				interval '1 month'
			) AS month_start
		)
		SELECT
			to_char(month_series.month_start, 'YYYY-MM') AS month,
			COALESCE(SUM(CASE WHEN m.mutation_type = 'masuk' THEN m.amount END), 0)::text AS total_incoming,
			COALESCE(SUM(CASE WHEN m.mutation_type = 'keluar' THEN m.amount END), 0)::text AS total_outgoing,
			(
				COALESCE(SUM(CASE WHEN m.mutation_type = 'masuk' THEN m.amount END), 0) -
				COALESCE(SUM(CASE WHEN m.mutation_type = 'keluar' THEN m.amount END), 0)
			)::text AS net_flow
		FROM month_series
		LEFT JOIN mutations m
			ON m.user_id = $1
			AND m.happened_at >= month_series.month_start
			AND m.happened_at < month_series.month_start + interval '1 month'
		GROUP BY month_series.month_start
		ORDER BY month_series.month_start ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer trendRows.Close()

	for trendRows.Next() {
		var item TrendPoint
		if err := trendRows.Scan(&item.Month, &item.TotalIncoming, &item.TotalOutgoing, &item.NetFlow); err != nil {
			return nil, err
		}
		overview.MonthlyTrend = append(overview.MonthlyTrend, item)
	}
	if err := trendRows.Err(); err != nil {
		return nil, err
	}

	categoryRows, err := r.pool.Query(ctx, `
		WITH total_outgoing AS (
			SELECT COALESCE(SUM(amount), 0) AS total
			FROM mutations
			WHERE user_id = $1
				AND mutation_type = 'keluar'
				AND happened_at >= date_trunc('month', CURRENT_DATE)
		)
		SELECT
			COALESCE(c.id::text, '') AS category_id,
			COALESCE(c.name, 'Uncategorized') AS category_name,
			COALESCE(SUM(m.amount)::text, '0') AS total_amount
		FROM mutations m
		LEFT JOIN categories c ON c.id = m.category_id
		WHERE m.user_id = $1
			AND m.mutation_type = 'keluar'
			AND m.happened_at >= date_trunc('month', CURRENT_DATE)
		GROUP BY c.id, c.name
		ORDER BY SUM(m.amount) DESC, category_name ASC
		LIMIT 5
	`, userID)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	totalOutgoing := parseFloat(overview.TotalOutgoing)
	for categoryRows.Next() {
		var item CategoryBreakdown
		if err := categoryRows.Scan(&item.CategoryID, &item.CategoryName, &item.TotalAmount); err != nil {
			return nil, err
		}
		if totalOutgoing > 0 {
			item.Share = formatShare(parseFloat(item.TotalAmount) / totalOutgoing)
		} else {
			item.Share = "0.00"
		}
		overview.OutgoingCategories = append(overview.OutgoingCategories, item)
	}
	if err := categoryRows.Err(); err != nil {
		return nil, err
	}

	return overview, nil
}

func parseFloat(value string) float64 {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return number
}

func formatShare(value float64) string {
	return fmt.Sprintf("%.2f", value*100)
}
