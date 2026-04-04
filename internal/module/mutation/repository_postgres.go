package mutation

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type walletState struct {
	ID        string
	UserID    string
	Balance   string
	IsActive  bool
	DeletedAt any
}

// PostgresRepository implements mutation persistence using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed mutation repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	wallet, err := r.lockWallet(ctx, tx, userID, input.WalletID)
	if err != nil {
		return nil, err
	}

	if err := r.applyWalletEffect(ctx, tx, wallet.ID, input.Type, input.Amount, false); err != nil {
		return nil, err
	}

	item := &Mutation{}
	err = tx.QueryRow(ctx, `
		INSERT INTO mutations (user_id, wallet_id, mutation_type, amount, description, related_to_debt, happened_at)
		VALUES ($1, $2, $3, $4::numeric, $5, FALSE, $6)
		RETURNING id, user_id, wallet_id, debt_id, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
	`, userID, input.WalletID, input.Type, input.Amount, input.Description, input.HappenedAt).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.DebtID,
		&item.Type,
		&item.Amount,
		&item.Description,
		&item.RelatedToDebt,
		&item.HappenedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *PostgresRepository) List(ctx context.Context, userID string) ([]Mutation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, wallet_id, debt_id, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
		FROM mutations
		WHERE user_id = $1
		ORDER BY happened_at DESC, created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Mutation
	for rows.Next() {
		var item Mutation
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.WalletID,
			&item.DebtID,
			&item.Type,
			&item.Amount,
			&item.Description,
			&item.RelatedToDebt,
			&item.HappenedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, userID, mutationID string) (*Mutation, error) {
	item := &Mutation{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, wallet_id, debt_id, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
		FROM mutations
		WHERE id = $1 AND user_id = $2
	`, mutationID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.DebtID,
		&item.Type,
		&item.Amount,
		&item.Description,
		&item.RelatedToDebt,
		&item.HappenedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMutationNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) Update(ctx context.Context, userID, mutationID string, input UpsertInput) (*Mutation, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	current, err := r.getByIDForUpdate(ctx, tx, userID, mutationID)
	if err != nil {
		return nil, err
	}

	if _, err := r.lockWallet(ctx, tx, userID, current.WalletID); err != nil {
		return nil, err
	}

	if err := r.applyWalletEffect(ctx, tx, current.WalletID, current.Type, current.Amount, true); err != nil {
		return nil, err
	}

	if input.WalletID != current.WalletID {
		if _, err := r.lockWallet(ctx, tx, userID, input.WalletID); err != nil {
			return nil, err
		}
	}

	if err := r.applyWalletEffect(ctx, tx, input.WalletID, input.Type, input.Amount, false); err != nil {
		return nil, err
	}

	item := &Mutation{}
	err = tx.QueryRow(ctx, `
		UPDATE mutations
		SET wallet_id = $3,
		    mutation_type = $4,
		    amount = $5::numeric,
		    description = $6,
		    happened_at = $7,
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, wallet_id, debt_id, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
	`, mutationID, userID, input.WalletID, input.Type, input.Amount, input.Description, input.HappenedAt).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.DebtID,
		&item.Type,
		&item.Amount,
		&item.Description,
		&item.RelatedToDebt,
		&item.HappenedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, userID, mutationID string) error {
	_, err := r.GetByID(ctx, userID, mutationID)
	if err != nil {
		return err
	}
	return ErrMutationDeleteNotAllowed
}

func (r *PostgresRepository) getByIDForUpdate(ctx context.Context, tx pgx.Tx, userID, mutationID string) (*Mutation, error) {
	item := &Mutation{}
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, wallet_id, debt_id, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
		FROM mutations
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, mutationID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.DebtID,
		&item.Type,
		&item.Amount,
		&item.Description,
		&item.RelatedToDebt,
		&item.HappenedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMutationNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) lockWallet(ctx context.Context, tx pgx.Tx, userID, walletID string) (*walletState, error) {
	wallet := &walletState{}
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, balance::text, is_active, deleted_at
		FROM wallets
		WHERE id = $1 AND user_id = $2 AND is_active = TRUE AND deleted_at IS NULL
		FOR UPDATE
	`, walletID, userID).Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.IsActive, &wallet.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMutationWalletNotFound
	}
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (r *PostgresRepository) applyWalletEffect(ctx context.Context, tx pgx.Tx, walletID, mutationType, amount string, reverse bool) error {
	switch mutationType {
	case "masuk":
		if reverse {
			_, err := tx.Exec(ctx, `UPDATE wallets SET balance = balance - $2::numeric, updated_at = NOW() WHERE id = $1`, walletID, amount)
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE wallets SET balance = balance + $2::numeric, updated_at = NOW() WHERE id = $1`, walletID, amount)
		return err
	case "keluar":
		if reverse {
			_, err := tx.Exec(ctx, `UPDATE wallets SET balance = balance + $2::numeric, updated_at = NOW() WHERE id = $1`, walletID, amount)
			return err
		}
		tag, err := tx.Exec(ctx, `
			UPDATE wallets
			SET balance = balance - $2::numeric, updated_at = NOW()
			WHERE id = $1 AND balance >= $2::numeric
		`, walletID, amount)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrInsufficientWalletBalance
		}
		return nil
	default:
		return ErrMutationValidation
	}
}
