package history

import (
	"context"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"

	"hnu-tickets/internal/platform/database"
	"hnu-tickets/internal/platform/fieldcrypt"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
	keys *fieldcrypt.Keyring
}

func NewPostgresRepository(pool *pgxpool.Pool, keys *fieldcrypt.Keyring) *PostgresRepository {
	return &PostgresRepository{pool: pool, keys: keys}
}

func (r *PostgresRepository) Create(ctx context.Context, entry *Entry) error {
	entry.ID = uuid.New()
	old, err := r.encryptValue(entry.ID, "old_value", entry.OldValue)
	if err != nil {
		return err
	}
	current, err := r.encryptValue(entry.ID, "new_value", entry.NewValue)
	if err != nil {
		return err
	}
	err = database.Executor(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO ticket_history (id, ticket_id, changed_by, action, old_value, new_value)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, entry.ID, entry.TicketID, entry.ChangedBy, entry.Action, old, current).Scan(&entry.ID, &entry.CreatedAt)
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
		for column, value := range map[string]**string{"old_value": &entry.OldValue, "new_value": &entry.NewValue} {
			if *value != nil {
				plain, err := r.keys.Decrypt(**value, fieldcrypt.Context("ticket_history", entry.ID.String(), column))
				if err != nil {
					return nil, err
				}
				*value = &plain
			}
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history: %w", err)
	}
	return entries, nil
}

func (r *PostgresRepository) encryptValue(id uuid.UUID, column string, value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	sealed, err := r.keys.Encrypt(*value, fieldcrypt.Context("ticket_history", id.String(), column))
	if err != nil {
		return nil, err
	}
	return &sealed, nil
}
