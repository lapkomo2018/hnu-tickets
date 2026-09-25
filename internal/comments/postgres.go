package comments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"hnu-tickets/internal/platform/database"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, comment *Comment) error {
	err := database.Executor(ctx, r.pool).QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO comments (ticket_id, author_id, content)
			VALUES ($1, $2, $3)
			RETURNING id, author_id, created_at
		)
		SELECT i.id, u.name, i.created_at
		FROM inserted i JOIN users u ON u.id = i.author_id
	`, comment.TicketID, comment.AuthorID, comment.Content).
		Scan(&comment.ID, &comment.AuthorName, &comment.CreatedAt)
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListByTicket(ctx context.Context, ticketID uuid.UUID) ([]Comment, error) {
	rows, err := database.Executor(ctx, r.pool).Query(ctx, `
		SELECT c.id, c.ticket_id, c.author_id, u.name, c.content, c.created_at
		FROM comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.ticket_id = $1
		ORDER BY c.created_at, c.id
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	result := make([]Comment, 0)
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.TicketID, &comment.AuthorID, &comment.AuthorName, &comment.Content, &comment.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		result = append(result, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}
	return result, nil
}
