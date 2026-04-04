package debt

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements debt persistence using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a debt repository backed by PostgreSQL.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, debt *Debt) error {
	query := `
		INSERT INTO debts (user_id, name, principal_amount, remaining_amount, tenor_value, tenor_unit, payment_amount, status, note)
		VALUES ($1, $2, $3::numeric, $3::numeric, $4, $5, $6::numeric, $7, $8)
		RETURNING id, principal_amount::text, remaining_amount::text, status, is_active, deleted_at, created_at, updated_at
	`

	return r.pool.QueryRow(
		ctx,
		query,
		debt.UserID,
		debt.Name,
		debt.PrincipalAmount,
		debt.TenorValue,
		debt.TenorUnit,
		debt.PaymentAmount,
		debt.Status,
		debt.Note,
	).Scan(
		&debt.ID,
		&debt.PrincipalAmount,
		&debt.RemainingAmount,
		&debt.Status,
		&debt.IsActive,
		&debt.DeletedAt,
		&debt.CreatedAt,
		&debt.UpdatedAt,
	)
}

func (r *PostgresRepository) List(ctx context.Context, userID string) ([]Debt, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, principal_amount::text, remaining_amount::text, tenor_value, tenor_unit, payment_amount::text, status, is_active, note, deleted_at, created_at, updated_at
		FROM debts
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Debt
	for rows.Next() {
		var item Debt
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Name,
			&item.PrincipalAmount,
			&item.RemainingAmount,
			&item.TenorValue,
			&item.TenorUnit,
			&item.PaymentAmount,
			&item.Status,
			&item.IsActive,
			&item.Note,
			&item.DeletedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, userID, debtID string) (*Debt, error) {
	item := &Debt{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, principal_amount::text, remaining_amount::text, tenor_value, tenor_unit, payment_amount::text, status, is_active, note, deleted_at, created_at, updated_at
		FROM debts
		WHERE id = $1 AND user_id = $2
	`, debtID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.Name,
		&item.PrincipalAmount,
		&item.RemainingAmount,
		&item.TenorValue,
		&item.TenorUnit,
		&item.PaymentAmount,
		&item.Status,
		&item.IsActive,
		&item.Note,
		&item.DeletedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDebtNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) Update(ctx context.Context, debt *Debt) (*Debt, error) {
	item := &Debt{}
	err := r.pool.QueryRow(ctx, `
		UPDATE debts
		SET name = $3,
		    tenor_value = $4,
		    tenor_unit = $5,
		    payment_amount = $6::numeric,
		    note = $7,
		    status = CASE WHEN remaining_amount = 0 THEN 'lunas' ELSE 'active' END,
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, principal_amount::text, remaining_amount::text, tenor_value, tenor_unit, payment_amount::text, status, is_active, note, deleted_at, created_at, updated_at
	`, debt.ID, debt.UserID, debt.Name, debt.TenorValue, debt.TenorUnit, debt.PaymentAmount, debt.Note).Scan(
		&item.ID,
		&item.UserID,
		&item.Name,
		&item.PrincipalAmount,
		&item.RemainingAmount,
		&item.TenorValue,
		&item.TenorUnit,
		&item.PaymentAmount,
		&item.Status,
		&item.IsActive,
		&item.Note,
		&item.DeletedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDebtNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) Inactivate(ctx context.Context, userID, debtID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE debts
		SET is_active = FALSE,
		    deleted_at = NOW(),
		    status = 'lunas',
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND remaining_amount = 0 AND deleted_at IS NULL
	`, debtID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	item, err := r.GetByID(ctx, userID, debtID)
	if err != nil {
		return err
	}
	if item.RemainingAmount != "0.00" && item.RemainingAmount != "0" {
		return ErrDebtRemainingNotZero
	}
	return nil
}
