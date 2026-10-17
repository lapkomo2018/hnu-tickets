package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"hnu-tickets/internal/platform/database"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, user *User) error {
	err := database.Executor(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, lower($2), $3, $4)
		RETURNING id, created_at, updated_at
	`, user.Name, user.Email, user.PasswordHash, user.Role).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if isUniqueViolation(err) {
		return ErrEmailAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	user.Email = strings.ToLower(user.Email)
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return scanUser(database.Executor(ctx, r.pool).QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users WHERE id = $1
	`, id))
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return scanUser(database.Executor(ctx, r.pool).QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users WHERE email = lower($1)
	`, email))
}

func (r *PostgresRepository) List(ctx context.Context) ([]User, error) {
	rows, err := database.Executor(ctx, r.pool).Query(ctx, `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users ORDER BY name, id
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	result := make([]User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) Update(ctx context.Context, user *User) error {
	err := database.Executor(ctx, r.pool).QueryRow(ctx, `
		UPDATE users SET name = $2, email = lower($3), updated_at = now()
		WHERE id = $1
		RETURNING updated_at
	`, user.ID, user.Name, user.Email).Scan(&user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	if isUniqueViolation(err) {
		return ErrEmailAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	user.Email = strings.ToLower(user.Email)
	return nil
}

func (r *PostgresRepository) UpdateRole(ctx context.Context, id uuid.UUID, role Role) error {
	command, err := database.Executor(ctx, r.pool).Exec(ctx, `
		UPDATE users SET role = $2, updated_at = now() WHERE id = $1
	`, id, role)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	if command.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

type scanner interface {
	Scan(...any) error
}

func scanUser(row scanner) (*User, error) {
	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
