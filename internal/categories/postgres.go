package categories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"hnu-tickets/internal/platform/apperror"
	"hnu-tickets/internal/platform/database"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, category *Category) error {
	err := database.Executor(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING id
	`, category.Name, category.Description).Scan(&category.ID)
	return categoryWriteError(err, "create")
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	var category Category
	err := database.Executor(ctx, r.pool).QueryRow(ctx, `
		SELECT id, name, description FROM categories WHERE id = $1
	`, id).Scan(&category.ID, &category.Name, &category.Description)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.New(apperror.KindNotFound, "CATEGORY_NOT_FOUND", "category not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &category, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]Category, error) {
	rows, err := database.Executor(ctx, r.pool).Query(ctx, `
		SELECT id, name, description FROM categories ORDER BY name, id
	`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	result := make([]Category, 0)
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		result = append(result, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) Update(ctx context.Context, category *Category) error {
	command, err := database.Executor(ctx, r.pool).Exec(ctx, `
		UPDATE categories SET name = $2, description = $3 WHERE id = $1
	`, category.ID, category.Name, category.Description)
	if err := categoryWriteError(err, "update"); err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return apperror.New(apperror.KindNotFound, "CATEGORY_NOT_FOUND", "category not found")
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	command, err := database.Executor(ctx, r.pool).Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return apperror.New(apperror.KindConflict, "CATEGORY_IN_USE", "category is used by one or more tickets")
	}
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if command.RowsAffected() == 0 {
		return apperror.New(apperror.KindNotFound, "CATEGORY_NOT_FOUND", "category not found")
	}
	return nil
}

func categoryWriteError(err error, operation string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.New(apperror.KindConflict, "CATEGORY_NAME_EXISTS", "a category with this name already exists")
	}
	if err != nil {
		return fmt.Errorf("%s category: %w", operation, err)
	}
	return nil
}
