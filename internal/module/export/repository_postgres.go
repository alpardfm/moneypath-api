package export

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements export read queries using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed export repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ListMutations(ctx context.Context, userID string, filter MutationFilter) ([]MutationRow, error) {
	whereParts := []string{"m.user_id = $1"}
	args := []any{userID}
	argIndex := 2

	if filter.Type != "" {
		whereParts = append(whereParts, fmt.Sprintf("m.mutation_type = $%d", argIndex))
		args = append(args, filter.Type)
		argIndex++
	}
	if filter.WalletID != "" {
		whereParts = append(whereParts, fmt.Sprintf("m.wallet_id = $%d", argIndex))
		args = append(args, filter.WalletID)
		argIndex++
	}
	if filter.CategoryID != "" {
		whereParts = append(whereParts, fmt.Sprintf("m.category_id = $%d", argIndex))
		args = append(args, filter.CategoryID)
		argIndex++
	}
	if filter.DebtID != "" {
		whereParts = append(whereParts, fmt.Sprintf("m.debt_id = $%d", argIndex))
		args = append(args, filter.DebtID)
		argIndex++
	}
	if filter.From != nil {
		whereParts = append(whereParts, fmt.Sprintf("m.happened_at >= $%d", argIndex))
		args = append(args, *filter.From)
		argIndex++
	}
	if filter.To != nil {
		whereParts = append(whereParts, fmt.Sprintf("m.happened_at <= $%d", argIndex))
		args = append(args, *filter.To)
		argIndex++
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT
			m.id,
			w.name,
			COALESCE(c.name, ''),
			COALESCE(d.name, ''),
			m.mutation_type,
			m.amount::text,
			m.description,
			m.related_to_debt,
			m.happened_at,
			m.created_at
		FROM mutations m
		INNER JOIN wallets w ON w.id = m.wallet_id
		LEFT JOIN categories c ON c.id = m.category_id
		LEFT JOIN debts d ON d.id = m.debt_id
		WHERE %s
		ORDER BY m.happened_at ASC, m.created_at ASC
	`, strings.Join(whereParts, " AND ")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MutationRow
	for rows.Next() {
		var item MutationRow
		if err := rows.Scan(
			&item.ID,
			&item.WalletName,
			&item.CategoryName,
			&item.DebtName,
			&item.Type,
			&item.Amount,
			&item.Description,
			&item.RelatedToDebt,
			&item.HappenedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
