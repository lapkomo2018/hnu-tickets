package analytics

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"hnu-tickets/internal/platform/database"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Overview(ctx context.Context, period Period) (Overview, error) {
	where, args := periodWhere(period, "created_at")
	var result Overview
	err := database.Executor(ctx, r.pool).QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status NOT IN ('CLOSED', 'CANCELLED')),
			COUNT(*) FILTER (WHERE status = 'IN_PROGRESS'),
			COUNT(*) FILTER (WHERE status = 'RESOLVED'),
			COUNT(*) FILTER (WHERE status = 'CLOSED'),
			COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))) FILTER (WHERE resolved_at IS NOT NULL), 0)
		FROM tickets`+where, args...).Scan(
		&result.TotalTickets, &result.OpenTickets, &result.InProgressTickets,
		&result.ResolvedTickets, &result.ClosedTickets, &result.AverageResolutionSeconds,
	)
	if err != nil {
		return Overview{}, fmt.Errorf("load analytics overview: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) ByStatus(ctx context.Context, period Period) ([]Count, error) {
	return r.groupedCounts(ctx, "status", period)
}

func (r *PostgresRepository) ByPriority(ctx context.Context, period Period) ([]Count, error) {
	return r.groupedCounts(ctx, "priority", period)
}

func (r *PostgresRepository) groupedCounts(ctx context.Context, column string, period Period) ([]Count, error) {
	where, args := periodWhere(period, "created_at")
	rows, err := database.Executor(ctx, r.pool).Query(ctx,
		fmt.Sprintf("SELECT %s, COUNT(*) FROM tickets%s GROUP BY %s ORDER BY %s", column, where, column, column), args...)
	if err != nil {
		return nil, fmt.Errorf("load analytics counts: %w", err)
	}
	defer rows.Close()

	result := make([]Count, 0)
	for rows.Next() {
		var item Count
		if err := rows.Scan(&item.Value, &item.Count); err != nil {
			return nil, fmt.Errorf("scan analytics count: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate analytics counts: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) ByCategory(ctx context.Context, period Period) ([]CategoryCount, error) {
	condition, args := periodJoinCondition(period, "t.created_at")
	rows, err := database.Executor(ctx, r.pool).Query(ctx, `
		SELECT c.id, c.name, COUNT(t.id)
		FROM categories c
		LEFT JOIN tickets t ON t.category_id = c.id`+condition+`
		GROUP BY c.id, c.name
		ORDER BY c.name, c.id
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("load category analytics: %w", err)
	}
	defer rows.Close()
	result := make([]CategoryCount, 0)
	for rows.Next() {
		var item CategoryCount
		if err := rows.Scan(&item.ID, &item.Name, &item.Count); err != nil {
			return nil, fmt.Errorf("scan category analytics: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate category analytics: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) Specialists(ctx context.Context, period Period) ([]SpecialistWorkload, error) {
	condition, args := periodJoinCondition(period, "t.created_at")
	rows, err := database.Executor(ctx, r.pool).Query(ctx, `
		SELECT u.id, u.name, COUNT(t.id)
		FROM users u
		LEFT JOIN tickets t ON t.assigned_to = u.id
			AND t.status IN ('ASSIGNED', 'IN_PROGRESS', 'WAITING_FOR_USER')`+condition+`
		WHERE u.role = 'SPECIALIST'
		GROUP BY u.id, u.name
		ORDER BY u.name, u.id
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("load specialist analytics: %w", err)
	}
	defer rows.Close()
	result := make([]SpecialistWorkload, 0)
	for rows.Next() {
		var item SpecialistWorkload
		if err := rows.Scan(&item.ID, &item.Name, &item.OpenTickets); err != nil {
			return nil, fmt.Errorf("scan specialist analytics: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate specialist analytics: %w", err)
	}
	return result, nil
}

func periodWhere(period Period, column string) (string, []any) {
	conditions, args := periodConditions(period, column)
	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func periodJoinCondition(period Period, column string) (string, []any) {
	conditions, args := periodConditions(period, column)
	if len(conditions) == 0 {
		return "", args
	}
	return " AND " + strings.Join(conditions, " AND "), args
}

func periodConditions(period Period, column string) ([]string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if period.From != nil {
		args = append(args, *period.From)
		conditions = append(conditions, fmt.Sprintf("%s >= $%d", column, len(args)))
	}
	if period.To != nil {
		args = append(args, *period.To)
		conditions = append(conditions, fmt.Sprintf("%s <= $%d", column, len(args)))
	}
	return conditions, args
}
