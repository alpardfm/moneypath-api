package healthscore

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements financial health scoring queries using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed financial health repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// GetSnapshot returns the aggregates needed to derive the financial health score.
func (r *PostgresRepository) GetSnapshot(ctx context.Context, userID string) (*Snapshot, error) {
	snapshot := &Snapshot{}
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT SUM(balance)::text FROM wallets WHERE user_id = $1 AND is_active = TRUE AND deleted_at IS NULL), '0'),
			COALESCE((SELECT SUM(remaining_amount)::text FROM debts WHERE user_id = $1 AND is_active = TRUE AND deleted_at IS NULL), '0'),
			COALESCE((
				SELECT SUM(amount)::text
				FROM mutations
				WHERE user_id = $1
					AND mutation_type = 'masuk'
					AND happened_at >= date_trunc('month', CURRENT_DATE) - interval '2 months'
			), '0'),
			COALESCE((
				SELECT SUM(amount)::text
				FROM mutations
				WHERE user_id = $1
					AND mutation_type = 'keluar'
					AND happened_at >= date_trunc('month', CURRENT_DATE) - interval '2 months'
			), '0')
	`, userID).Scan(
		&snapshot.TotalAssets,
		&snapshot.TotalDebts,
		&snapshot.RecentIncoming,
		&snapshot.RecentOutgoing,
	)
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}
