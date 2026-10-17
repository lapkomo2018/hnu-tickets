package history

import (
	"context"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"

	"hnu-tickets/internal/platform/database"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, entry *Entry) error {
	err := database.Executor(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO ticket_history (ticket_id, changed_by, action, old_value, new_value)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, entry.TicketID, entry.ChangedBy, entry.Action, entry.OldValue, entry.NewValue).Scan(&entry.ID, &entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("create history entry: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListByTicket(ctx context.Context, ticketID uuid.UUID) ([]Entry, error) {
	rows, err := database.Executor(ctx, r.pool).Query(ctx, `
		SELECT id, ticket_id, changed_by, action, old_value, new_value, created_at
		FROM ticket_history
		WHERE ticket_id = $1
		ORDER BY created_at, id
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.TicketID, &entry.ChangedBy, &entry.Action, &entry.OldValue, &entry.NewValue, &entry.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan history entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history: %w", err)
	}
	return entries, nil
}
