package recurring

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

// PostgresRepository implements recurring persistence using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed recurring repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, rule *Rule) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO recurring_rules (
			user_id, wallet_id, category_id, mutation_type, amount, description,
			interval_unit, interval_step, start_at, end_at, next_run_at
		)
		VALUES ($1, $2, $3, $4, $5::numeric, $6, $7, $8, $9, $10, $11)
		RETURNING id, last_run_at, is_active, deleted_at, created_at, updated_at
	`, rule.UserID, rule.WalletID, rule.CategoryID, rule.Type, rule.Amount, rule.Description, rule.IntervalUnit, rule.IntervalStep, rule.StartAt, rule.EndAt, rule.NextRunAt).Scan(
		&rule.ID,
		&rule.LastRunAt,
		&rule.IsActive,
		&rule.DeletedAt,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRuleValidation
	}
	return err
}

func (r *PostgresRepository) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	whereParts := []string{"user_id = $1", "deleted_at IS NULL", "is_active = TRUE"}
	args := []any{userID}
	argIndex := 2

	if options.Type != "" {
		whereParts = append(whereParts, fmt.Sprintf("mutation_type = $%d", argIndex))
		args = append(args, options.Type)
		argIndex++
	}

	whereClause := strings.Join(whereParts, " AND ")
	var totalItems int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(1) FROM recurring_rules WHERE %s`, whereClause), args...).Scan(&totalItems); err != nil {
		return nil, err
	}

	offset := (options.Page - 1) * options.PageSize
	args = append(args, options.PageSize, offset)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, user_id, wallet_id, category_id, mutation_type, amount::text, description, interval_unit, interval_step, start_at, end_at, next_run_at, last_run_at, is_active, deleted_at, created_at, updated_at
		FROM recurring_rules
		WHERE %s
		ORDER BY next_run_at ASC, created_at ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Rule
	for rows.Next() {
		var item Rule
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.WalletID,
			&item.CategoryID,
			&item.Type,
			&item.Amount,
			&item.Description,
			&item.IntervalUnit,
			&item.IntervalStep,
			&item.StartAt,
			&item.EndAt,
			&item.NextRunAt,
			&item.LastRunAt,
			&item.IsActive,
			&item.DeletedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return &ListResult{Items: items, TotalItems: totalItems}, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, userID, ruleID string) (*Rule, error) {
	item := &Rule{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, wallet_id, category_id, mutation_type, amount::text, description, interval_unit, interval_step, start_at, end_at, next_run_at, last_run_at, is_active, deleted_at, created_at, updated_at
		FROM recurring_rules
		WHERE id = $1 AND user_id = $2
	`, ruleID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.CategoryID,
		&item.Type,
		&item.Amount,
		&item.Description,
		&item.IntervalUnit,
		&item.IntervalStep,
		&item.StartAt,
		&item.EndAt,
		&item.NextRunAt,
		&item.LastRunAt,
		&item.IsActive,
		&item.DeletedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) Update(ctx context.Context, rule *Rule) (*Rule, error) {
	item := &Rule{}
	err := r.pool.QueryRow(ctx, `
		UPDATE recurring_rules
		SET wallet_id = $3,
		    category_id = $4,
		    mutation_type = $5,
		    amount = $6::numeric,
		    description = $7,
		    interval_unit = $8,
		    interval_step = $9,
		    start_at = $10,
		    end_at = $11,
		    next_run_at = $10,
		    last_run_at = NULL,
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, wallet_id, category_id, mutation_type, amount::text, description, interval_unit, interval_step, start_at, end_at, next_run_at, last_run_at, is_active, deleted_at, created_at, updated_at
	`, rule.ID, rule.UserID, rule.WalletID, rule.CategoryID, rule.Type, rule.Amount, rule.Description, rule.IntervalUnit, rule.IntervalStep, rule.StartAt, rule.EndAt).Scan(
		&item.ID,
		&item.UserID,
		&item.WalletID,
		&item.CategoryID,
		&item.Type,
		&item.Amount,
		&item.Description,
		&item.IntervalUnit,
		&item.IntervalStep,
		&item.StartAt,
		&item.EndAt,
		&item.NextRunAt,
		&item.LastRunAt,
		&item.IsActive,
		&item.DeletedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) Inactivate(ctx context.Context, userID, ruleID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE recurring_rules
		SET is_active = FALSE, deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, ruleID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRuleNotFound
	}
	return nil
}

func (r *PostgresRepository) RunDue(ctx context.Context, userID string, now time.Time) (*RunDueResult, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, user_id, wallet_id, category_id, mutation_type, amount::text, description, interval_unit, interval_step, start_at, end_at, next_run_at, last_run_at, is_active, deleted_at, created_at, updated_at
		FROM recurring_rules
		WHERE user_id = $1
		  AND is_active = TRUE
		  AND deleted_at IS NULL
		  AND next_run_at <= $2
		  AND (end_at IS NULL OR next_run_at <= end_at)
		ORDER BY next_run_at ASC, created_at ASC
		FOR UPDATE
	`, userID, now)
	if err != nil {
		return nil, err
	}

	result := &RunDueResult{}
	var dueRules []Rule
	for rows.Next() {
		var rule Rule
		if err := rows.Scan(
			&rule.ID,
			&rule.UserID,
			&rule.WalletID,
			&rule.CategoryID,
			&rule.Type,
			&rule.Amount,
			&rule.Description,
			&rule.IntervalUnit,
			&rule.IntervalStep,
			&rule.StartAt,
			&rule.EndAt,
			&rule.NextRunAt,
			&rule.LastRunAt,
			&rule.IsActive,
			&rule.DeletedAt,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			rows.Close()
			return nil, err
		}
		dueRules = append(dueRules, rule)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for i := range dueRules {
		rule := dueRules[i]
		if err := r.executeRule(ctx, tx, &rule); err != nil {
			result.Skipped = append(result.Skipped, SkippedRunItem{RuleID: rule.ID, Reason: err.Error()})
			continue
		}
		result.Processed++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresRepository) executeRule(ctx context.Context, tx pgx.Tx, rule *Rule) error {
	if _, err := r.lockWallet(ctx, tx, rule.UserID, rule.WalletID); err != nil {
		return err
	}
	if rule.CategoryID != nil {
		category, err := r.lockCategory(ctx, tx, rule.UserID, *rule.CategoryID)
		if err != nil {
			return err
		}
		if category.Type != rule.Type {
			return ErrRuleCategoryMismatch
		}
	}
	if err := r.applyWalletEffect(ctx, tx, rule.WalletID, rule.Type, rule.Amount); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO mutations (user_id, wallet_id, category_id, debt_id, debt_action, mutation_type, amount, description, related_to_debt, happened_at)
		VALUES ($1, $2, $3, NULL, 'none', $4, $5::numeric, $6, FALSE, $7)
	`, rule.UserID, rule.WalletID, rule.CategoryID, rule.Type, rule.Amount, rule.Description, rule.NextRunAt); err != nil {
		return err
	}

	nextRunAt := nextRunAt(rule.NextRunAt, rule.IntervalUnit, rule.IntervalStep)
	isActive := true
	if rule.EndAt != nil && nextRunAt.After(*rule.EndAt) {
		isActive = false
	}

	_, err := tx.Exec(ctx, `
		UPDATE recurring_rules
		SET last_run_at = $3,
		    next_run_at = $4,
		    is_active = $5,
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, rule.ID, rule.UserID, rule.NextRunAt, nextRunAt, isActive)
	return err
}

func (r *PostgresRepository) lockWallet(ctx context.Context, tx pgx.Tx, userID, walletID string) (*walletState, error) {
	item := &walletState{}
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, balance::text, is_active, deleted_at
		FROM wallets
		WHERE id = $1 AND user_id = $2 AND is_active = TRUE AND deleted_at IS NULL
		FOR UPDATE
	`, walletID, userID).Scan(&item.ID, &item.UserID, &item.Balance, &item.IsActive, &item.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleWalletNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
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
		return nil, ErrRuleCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) applyWalletEffect(ctx context.Context, tx pgx.Tx, walletID, mutationType, amount string) error {
	switch mutationType {
	case "masuk":
		_, err := tx.Exec(ctx, `UPDATE wallets SET balance = balance + $2::numeric, updated_at = NOW() WHERE id = $1`, walletID, amount)
		return err
	case "keluar":
		tag, err := tx.Exec(ctx, `
			UPDATE wallets
			SET balance = balance - $2::numeric, updated_at = NOW()
			WHERE id = $1 AND balance >= $2::numeric
		`, walletID, amount)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrRuleValidation
		}
		return nil
	default:
		return ErrRuleValidation
	}
}

func nextRunAt(current time.Time, intervalUnit string, intervalStep int) time.Time {
	switch intervalUnit {
	case "daily":
		return current.AddDate(0, 0, intervalStep)
	case "weekly":
		return current.AddDate(0, 0, 7*intervalStep)
	default:
		return current.AddDate(0, intervalStep, 0)
	}
}
