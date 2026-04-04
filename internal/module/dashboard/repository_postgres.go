package dashboard

import (
	"context"

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
	return overview, rows.Err()
}
