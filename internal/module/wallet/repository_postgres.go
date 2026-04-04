package wallet

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements wallet persistence using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a wallet repository backed by PostgreSQL.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, wallet *Wallet) error {
	query := `
		INSERT INTO wallets (user_id, name)
		VALUES ($1, $2)
		RETURNING id, balance::text, is_active, deleted_at, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, wallet.UserID, wallet.Name).
		Scan(&wallet.ID, &wallet.Balance, &wallet.IsActive, &wallet.DeletedAt, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err == nil {
		return nil
	}
	return mapConstraintError(err)
}

func (r *PostgresRepository) ListActive(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	var totalItems int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM wallets
		WHERE user_id = $1 AND is_active = TRUE AND deleted_at IS NULL
	`, userID).Scan(&totalItems); err != nil {
		return nil, err
	}

	offset := (options.Page - 1) * options.PageSize
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, balance::text, is_active, deleted_at, created_at, updated_at
		FROM wallets
		WHERE user_id = $1 AND is_active = TRUE AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`, userID, options.PageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []Wallet
	for rows.Next() {
		var wallet Wallet
		if err := rows.Scan(&wallet.ID, &wallet.UserID, &wallet.Name, &wallet.Balance, &wallet.IsActive, &wallet.DeletedAt, &wallet.CreatedAt, &wallet.UpdatedAt); err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}

	return &ListResult{
		Items:      wallets,
		TotalItems: totalItems,
	}, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, userID, walletID string) (*Wallet, error) {
	wallet := &Wallet{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, balance::text, is_active, deleted_at, created_at, updated_at
		FROM wallets
		WHERE id = $1 AND user_id = $2
	`, walletID, userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Balance,
		&wallet.IsActive,
		&wallet.DeletedAt,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWalletNotFound
	}
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (r *PostgresRepository) UpdateName(ctx context.Context, userID, walletID, name string) (*Wallet, error) {
	wallet := &Wallet{}
	err := r.pool.QueryRow(ctx, `
		UPDATE wallets
		SET name = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, balance::text, is_active, deleted_at, created_at, updated_at
	`, walletID, userID, name).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Balance,
		&wallet.IsActive,
		&wallet.DeletedAt,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWalletNotFound
	}
	if err != nil {
		return nil, mapConstraintError(err)
	}
	return wallet, nil
}

func (r *PostgresRepository) Inactivate(ctx context.Context, userID, walletID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE wallets
		SET is_active = FALSE, deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND balance = 0 AND deleted_at IS NULL
	`, walletID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	wallet, err := r.GetByID(ctx, userID, walletID)
	if err != nil {
		return err
	}
	if wallet.Balance != "0.00" && wallet.Balance != "0" {
		return ErrWalletBalanceNotZero
	}

	return nil
}

func mapConstraintError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "wallets_user_id_name_key") {
			return ErrWalletNameAlreadyUsed
		}
	}
	return err
}
