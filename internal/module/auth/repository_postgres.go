package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements the auth repository using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a PostgreSQL-backed auth repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (email, username, password_hash, full_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, preferred_currency, timezone, date_format, week_start_day, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, user.Email, user.Username, user.PasswordHash, user.FullName).
		Scan(&user.ID, &user.PreferredCurrency, &user.Timezone, &user.DateFormat, &user.WeekStartDay, &user.CreatedAt, &user.UpdatedAt)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "email") {
			return ErrEmailAlreadyUsed
		}
		if strings.Contains(pgErr.ConstraintName, "username") {
			return ErrUsernameAlreadyUsed
		}
	}

	return err
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return r.getOne(ctx, `SELECT id, email, username, password_hash, full_name, preferred_currency, timezone, date_format, week_start_day, created_at, updated_at FROM users WHERE email = $1`, email)
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return r.getOne(ctx, `SELECT id, email, username, password_hash, full_name, preferred_currency, timezone, date_format, week_start_day, created_at, updated_at FROM users WHERE username = $1`, username)
}

func (r *PostgresRepository) GetUserByEmailOrUsername(ctx context.Context, value string) (*User, error) {
	return r.getOne(ctx, `SELECT id, email, username, password_hash, full_name, preferred_currency, timezone, date_format, week_start_day, created_at, updated_at FROM users WHERE email = $1 OR username = $1`, value)
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	return r.getOne(ctx, `SELECT id, email, username, password_hash, full_name, preferred_currency, timezone, date_format, week_start_day, created_at, updated_at FROM users WHERE id = $1`, userID)
}

func (r *PostgresRepository) UpdateProfile(ctx context.Context, userID, email, username, fullName string) (*User, error) {
	query := `
		UPDATE users
		SET email = $2, username = $3, full_name = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, username, password_hash, full_name, preferred_currency, timezone, date_format, week_start_day, created_at, updated_at
	`

	user := &User{}
	err := r.pool.QueryRow(ctx, query, userID, email, username, fullName).
		Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.FullName, &user.PreferredCurrency, &user.Timezone, &user.DateFormat, &user.WeekStartDay, &user.CreatedAt, &user.UpdatedAt)
	if err == nil {
		return user, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "email") {
			return nil, ErrEmailAlreadyUsed
		}
		if strings.Contains(pgErr.ConstraintName, "username") {
			return nil, ErrUsernameAlreadyUsed
		}
	}

	return nil, err
}

func (r *PostgresRepository) UpdateSettings(ctx context.Context, userID, preferredCurrency, timezone, dateFormat, weekStartDay string) (*User, error) {
	user := &User{}
	err := r.pool.QueryRow(ctx, `
		UPDATE users
		SET preferred_currency = $2,
		    timezone = $3,
		    date_format = $4,
		    week_start_day = $5,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, username, password_hash, full_name, preferred_currency, timezone, date_format, week_start_day, created_at, updated_at
	`, userID, preferredCurrency, timezone, dateFormat, weekStartDay).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.FullName,
		&user.PreferredCurrency, &user.Timezone, &user.DateFormat, &user.WeekStartDay,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`, userID, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresRepository) getOne(ctx context.Context, query string, arg string) (*User, error) {
	user := &User{}
	err := r.pool.QueryRow(ctx, query, arg).
		Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.FullName, &user.PreferredCurrency, &user.Timezone, &user.DateFormat, &user.WeekStartDay, &user.CreatedAt, &user.UpdatedAt)
	if err == nil {
		return user, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return nil, err
}
