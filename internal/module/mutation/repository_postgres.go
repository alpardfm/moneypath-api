package mutation

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

type categoryState struct {
	ID        string
	UserID    string
	Type      string
	IsActive  bool
	DeletedAt any
}

type debtState struct {
	ID              string
	UserID          string
	PrincipalAmount string
	RemainingAmount string
	Status          string
	IsActive        bool
	DeletedAt       any
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

	if _, err := r.lockWallet(ctx, tx, userID, input.WalletID); err != nil {
		return nil, err
	}

	mutationDebtID, debtAction, err := r.applyNewState(ctx, tx, userID, input)
	if err != nil {
		return nil, err
	}

	item := &Mutation{}
	err = tx.QueryRow(ctx, `
		INSERT INTO mutations (user_id, wallet_id, category_id, debt_id, debt_action, mutation_type, amount, description, related_to_debt, happened_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::numeric, $8, $9, $10)
		RETURNING id, user_id, wallet_id, category_id, debt_id, debt_action, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
	`, userID, input.WalletID, input.CategoryID, mutationDebtID, debtAction, input.Type, input.Amount, input.Description, input.RelatedToDebt, input.HappenedAt).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.CategoryID,
		&item.DebtID,
		&item.DebtAction,
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

func (r *PostgresRepository) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	whereParts := []string{"user_id = $1"}
	args := []any{userID}
	argIndex := 2

	if options.Type != "" {
		whereParts = append(whereParts, fmt.Sprintf("mutation_type = $%d", argIndex))
		args = append(args, options.Type)
		argIndex++
	}
	if options.WalletID != "" {
		whereParts = append(whereParts, fmt.Sprintf("wallet_id = $%d", argIndex))
		args = append(args, options.WalletID)
		argIndex++
	}
	if options.CategoryID != "" {
		whereParts = append(whereParts, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, options.CategoryID)
		argIndex++
	}
	if options.DebtID != "" {
		whereParts = append(whereParts, fmt.Sprintf("debt_id = $%d", argIndex))
		args = append(args, options.DebtID)
		argIndex++
	}
	if options.RelatedToDebt != nil {
		whereParts = append(whereParts, fmt.Sprintf("related_to_debt = $%d", argIndex))
		args = append(args, *options.RelatedToDebt)
		argIndex++
	}
	if options.From != "" {
		whereParts = append(whereParts, fmt.Sprintf("happened_at >= $%d::timestamptz", argIndex))
		args = append(args, options.From)
		argIndex++
	}
	if options.To != "" {
		whereParts = append(whereParts, fmt.Sprintf("happened_at <= $%d::timestamptz", argIndex))
		args = append(args, options.To)
		argIndex++
	}

	whereClause := strings.Join(whereParts, " AND ")
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM mutations WHERE %s", whereClause)

	var totalItems int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalItems); err != nil {
		return nil, err
	}

	orderBy := "happened_at DESC, created_at DESC"
	switch options.SortBy {
	case "created_at":
		orderBy = fmt.Sprintf("created_at %s, happened_at %s", strings.ToUpper(options.SortDirection), strings.ToUpper(options.SortDirection))
	case "amount":
		orderBy = fmt.Sprintf("amount %s, created_at DESC", strings.ToUpper(options.SortDirection))
	default:
		orderBy = fmt.Sprintf("happened_at %s, created_at %s", strings.ToUpper(options.SortDirection), strings.ToUpper(options.SortDirection))
	}

	offset := (options.Page - 1) * options.PageSize
	args = append(args, options.PageSize, offset)
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, wallet_id, category_id, debt_id, debt_action, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
		FROM mutations
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
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
			&item.CategoryID,
			&item.DebtID,
			&item.DebtAction,
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
	return &ListResult{
		Items:      items,
		TotalItems: totalItems,
	}, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, userID, mutationID string) (*Mutation, error) {
	item := &Mutation{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, wallet_id, category_id, debt_id, debt_action, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
		FROM mutations
		WHERE id = $1 AND user_id = $2
	`, mutationID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.CategoryID,
		&item.DebtID,
		&item.DebtAction,
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
	if input.WalletID != current.WalletID {
		if _, err := r.lockWallet(ctx, tx, userID, input.WalletID); err != nil {
			return nil, err
		}
	}

	if err := r.reverseCurrentState(ctx, tx, userID, current); err != nil {
		return nil, err
	}

	mutationDebtID, debtAction, err := r.applyNewState(ctx, tx, userID, input)
	if err != nil {
		return nil, err
	}

	item := &Mutation{}
	err = tx.QueryRow(ctx, `
		UPDATE mutations
		SET wallet_id = $3,
		    category_id = $4,
		    debt_id = $5,
		    debt_action = $6,
		    mutation_type = $7,
		    amount = $8::numeric,
		    description = $9,
		    related_to_debt = $10,
		    happened_at = $11,
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, wallet_id, category_id, debt_id, debt_action, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
	`, mutationID, userID, input.WalletID, input.CategoryID, mutationDebtID, debtAction, input.Type, input.Amount, input.Description, input.RelatedToDebt, input.HappenedAt).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.CategoryID,
		&item.DebtID,
		&item.DebtAction,
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
		SELECT id, user_id, wallet_id, category_id, debt_id, debt_action, mutation_type, amount::text, description, related_to_debt, happened_at, created_at, updated_at
		FROM mutations
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, mutationID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.CategoryID,
		&item.DebtID,
		&item.DebtAction,
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

func (r *PostgresRepository) lockCategory(ctx context.Context, tx pgx.Tx, userID, categoryID string) (*categoryState, error) {
	item := &categoryState{}
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, category_type, is_active, deleted_at
		FROM categories
		WHERE id = $1 AND user_id = $2 AND is_active = TRUE AND deleted_at IS NULL
		FOR SHARE
	`, categoryID, userID).Scan(&item.ID, &item.UserID, &item.Type, &item.IsActive, &item.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMutationCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) lockDebt(ctx context.Context, tx pgx.Tx, userID, debtID string) (*debtState, error) {
	item := &debtState{}
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, principal_amount::text, remaining_amount::text, status, is_active, deleted_at
		FROM debts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, debtID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.PrincipalAmount,
		&item.RemainingAmount,
		&item.Status,
		&item.IsActive,
		&item.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMutationDebtNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) applyNewState(ctx context.Context, tx pgx.Tx, userID string, input UpsertInput) (*string, string, error) {
	if input.CategoryID != nil {
		category, err := r.lockCategory(ctx, tx, userID, *input.CategoryID)
		if err != nil {
			return nil, "", err
		}
		if category.Type != input.Type {
			return nil, "", ErrMutationCategoryMismatch
		}
	}

	if err := r.applyWalletEffect(ctx, tx, input.WalletID, input.Type, input.Amount, false); err != nil {
		return nil, "", err
	}

	if !input.RelatedToDebt {
		return nil, "none", nil
	}

	if input.Type == "keluar" {
		debt, err := r.lockDebt(ctx, tx, userID, *input.DebtID)
		if err != nil {
			return nil, "", err
		}
		if err := r.applyDebtDelta(ctx, tx, debt.ID, input.Amount, true); err != nil {
			return nil, "", err
		}
		return input.DebtID, "payment", nil
	}

	if input.DebtID != nil {
		debt, err := r.lockDebt(ctx, tx, userID, *input.DebtID)
		if err != nil {
			return nil, "", err
		}
		if err := r.applyDebtDelta(ctx, tx, debt.ID, input.Amount, false); err != nil {
			return nil, "", err
		}
		return input.DebtID, "borrow_existing", nil
	}

	debtID, err := r.createDebtFromMutation(ctx, tx, userID, input.NewDebt)
	if err != nil {
		return nil, "", err
	}
	return &debtID, "borrow_new", nil
}

func (r *PostgresRepository) reverseCurrentState(ctx context.Context, tx pgx.Tx, userID string, current *Mutation) error {
	if err := r.applyWalletEffect(ctx, tx, current.WalletID, current.Type, current.Amount, true); err != nil {
		return err
	}

	if !current.RelatedToDebt || current.DebtID == nil {
		return nil
	}

	switch current.DebtAction {
	case "payment":
		return r.applyDebtDelta(ctx, tx, *current.DebtID, current.Amount, false)
	case "borrow_existing":
		return r.reverseBorrowExisting(ctx, tx, userID, *current.DebtID, current.Amount)
	case "borrow_new":
		return r.reverseBorrowNew(ctx, tx, userID, current.ID, *current.DebtID)
	default:
		return nil
	}
}

func (r *PostgresRepository) reverseBorrowExisting(ctx context.Context, tx pgx.Tx, userID, debtID, amount string) error {
	debt, err := r.lockDebt(ctx, tx, userID, debtID)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE debts
		SET remaining_amount = remaining_amount - $2::numeric,
		    status = CASE WHEN remaining_amount - $2::numeric = 0 THEN 'lunas' ELSE 'active' END,
		    updated_at = NOW()
		WHERE id = $1 AND remaining_amount >= $2::numeric
	`, debt.ID, amount)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrDebtStateChanged
	}
	return nil
}

func (r *PostgresRepository) reverseBorrowNew(ctx context.Context, tx pgx.Tx, userID, mutationID, debtID string) error {
	var otherCount int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM mutations
		WHERE debt_id = $1 AND id <> $2
	`, debtID, mutationID).Scan(&otherCount)
	if err != nil {
		return err
	}
	if otherCount > 0 {
		return ErrDebtStateChanged
	}

	tag, err := tx.Exec(ctx, `DELETE FROM debts WHERE id = $1 AND user_id = $2`, debtID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrMutationDebtNotFound
	}
	return nil
}

func (r *PostgresRepository) createDebtFromMutation(ctx context.Context, tx pgx.Tx, userID string, input *NewDebtInput) (string, error) {
	var debtID string
	err := tx.QueryRow(ctx, `
		INSERT INTO debts (user_id, name, principal_amount, remaining_amount, tenor_value, tenor_unit, payment_amount, status, note)
		VALUES ($1, $2, $3::numeric, $3::numeric, $4, $5, $6::numeric, 'active', $7)
		RETURNING id
	`, userID, input.Name, input.Principal, input.TenorValue, input.TenorUnit, input.PaymentAmount, input.Note).Scan(&debtID)
	if err != nil {
		return "", err
	}
	return debtID, nil
}

func (r *PostgresRepository) applyDebtDelta(ctx context.Context, tx pgx.Tx, debtID, amount string, decrease bool) error {
	if decrease {
		tag, err := tx.Exec(ctx, `
			UPDATE debts
			SET remaining_amount = remaining_amount - $2::numeric,
			    status = CASE WHEN remaining_amount - $2::numeric = 0 THEN 'lunas' ELSE 'active' END,
			    updated_at = NOW()
			WHERE id = $1 AND remaining_amount >= $2::numeric
		`, debtID, amount)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrDebtStateChanged
		}
		return nil
	}

	_, err := tx.Exec(ctx, `
		UPDATE debts
		SET remaining_amount = remaining_amount + $2::numeric,
		    status = 'active',
		    is_active = TRUE,
		    deleted_at = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`, debtID, amount)
	return err
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

func ptrString(value string) *string {
	v := strings.TrimSpace(value)
	return &v
}
