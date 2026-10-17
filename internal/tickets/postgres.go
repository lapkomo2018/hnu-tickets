package tickets

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
	"hnu-tickets/internal/platform/fieldcrypt"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
	keys *fieldcrypt.Keyring
}

func NewPostgresRepository(pool *pgxpool.Pool, keys *fieldcrypt.Keyring) *PostgresRepository {
	return &PostgresRepository{pool: pool, keys: keys}
}

func (r *PostgresRepository) Create(ctx context.Context, ticket *Ticket) error {
	ticket.ID = uuid.New()
	sealed, err := r.keys.Encrypt(ticket.Description, fieldcrypt.Context("tickets", ticket.ID.String(), "description"))
	if err != nil {
		return err
	}
	err = database.Executor(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO tickets (id, title, description, status, priority, category_id, created_by, assigned_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, ticket.ID, ticket.Title, sealed, ticket.Status, ticket.Priority, ticket.CategoryID, ticket.CreatedBy, ticket.AssignedTo).
		Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)
	if isForeignKeyViolation(err) {
		return ErrInvalidCategory
	}
	if err != nil {
		return fmt.Errorf("create ticket: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Ticket, error) {
	return r.scanTicket(database.Executor(ctx, r.pool).QueryRow(ctx, ticketSelect+` WHERE id = $1`, id))
}

func (r *PostgresRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*Ticket, error) {
	return r.scanTicket(database.Executor(ctx, r.pool).QueryRow(ctx, ticketSelect+` WHERE id = $1 FOR UPDATE`, id))
}

func (r *PostgresRepository) GetRecord(ctx context.Context, id uuid.UUID) (*Record, error) {
	return r.scanRecord(database.Executor(ctx, r.pool).QueryRow(ctx, recordSelect+` WHERE t.id = $1`, id))
}

func (r *PostgresRepository) List(ctx context.Context, filter Filter) ([]Record, int64, error) {
	conditions, args := buildConditions(filter)
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := database.Executor(ctx, r.pool).QueryRow(ctx, `SELECT COUNT(*) FROM tickets t`+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tickets: %w", err)
	}

	sortColumns := map[string]string{
		"createdAt": "t.created_at",
		"updatedAt": "t.updated_at",
		"priority":  "t.priority",
		"status":    "t.status",
	}
	sortColumn := sortColumns[filter.Sort]
	if sortColumn == "" {
		sortColumn = "t.created_at"
	}
	order := "DESC"
	if strings.EqualFold(filter.Order, "asc") {
		order = "ASC"
	}

	limitPosition := len(args) + 1
	query := recordSelect + where + fmt.Sprintf(" ORDER BY %s %s, t.id %s LIMIT $%d OFFSET $%d", sortColumn, order, order, limitPosition, limitPosition+1)
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := database.Executor(ctx, r.pool).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	result := make([]Record, 0)
	for rows.Next() {
		record, err := r.scanRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate tickets: %w", err)
	}
	return result, total, nil
}

func (r *PostgresRepository) Update(ctx context.Context, ticket *Ticket) error {
	sealed, err := r.keys.Encrypt(ticket.Description, fieldcrypt.Context("tickets", ticket.ID.String(), "description"))
	if err != nil {
		return err
	}
	command, err := database.Executor(ctx, r.pool).Exec(ctx, `
		UPDATE tickets SET
			title = $2, description = $3, status = $4, priority = $5,
			category_id = $6, assigned_to = $7, updated_at = $8,
			resolved_at = $9, closed_at = $10
		WHERE id = $1
	`, ticket.ID, ticket.Title, sealed, ticket.Status, ticket.Priority,
		ticket.CategoryID, ticket.AssignedTo, ticket.UpdatedAt, ticket.ResolvedAt, ticket.ClosedAt)
	if isForeignKeyViolation(err) {
		return ticketReferenceError(err)
	}
	if err != nil {
		return fmt.Errorf("update ticket: %w", err)
	}
	if command.RowsAffected() == 0 {
		return ticketNotFound()
	}
	return nil
}

const ticketSelect = `
	SELECT id, title, description, status, priority, category_id, created_by,
	       assigned_to, created_at, updated_at, resolved_at, closed_at
	FROM tickets`

const recordSelect = `
	SELECT t.id, t.title, t.description, t.status, t.priority, t.category_id, t.created_by,
	       t.assigned_to, t.created_at, t.updated_at, t.resolved_at, t.closed_at,
	       c.id, c.name, creator.id, creator.name, assignee.name
	FROM tickets t
	JOIN categories c ON c.id = t.category_id
	JOIN users creator ON creator.id = t.created_by
	LEFT JOIN users assignee ON assignee.id = t.assigned_to`

type scanner interface {
	Scan(...any) error
}

func (r *PostgresRepository) scanTicket(row scanner) (*Ticket, error) {
	var ticket Ticket
	err := row.Scan(
		&ticket.ID, &ticket.Title, &ticket.Description, &ticket.Status, &ticket.Priority,
		&ticket.CategoryID, &ticket.CreatedBy, &ticket.AssignedTo, &ticket.CreatedAt,
		&ticket.UpdatedAt, &ticket.ResolvedAt, &ticket.ClosedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ticketNotFound()
	}
	if err != nil {
		return nil, fmt.Errorf("scan ticket: %w", err)
	}
	ticket.Description, err = r.keys.Decrypt(ticket.Description, fieldcrypt.Context("tickets", ticket.ID.String(), "description"))
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *PostgresRepository) scanRecord(row scanner) (*Record, error) {
	var record Record
	err := row.Scan(
		&record.ID, &record.Title, &record.Description, &record.Status, &record.Priority,
		&record.CategoryID, &record.CreatedBy, &record.AssignedTo, &record.CreatedAt,
		&record.UpdatedAt, &record.ResolvedAt, &record.ClosedAt,
		&record.Category.ID, &record.Category.Name, &record.Creator.ID, &record.Creator.Name,
		&record.AssigneeName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ticketNotFound()
	}
	if err != nil {
		return nil, fmt.Errorf("scan ticket record: %w", err)
	}
	record.Description, err = r.keys.Decrypt(record.Description, fieldcrypt.Context("tickets", record.ID.String(), "description"))
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func buildConditions(filter Filter) ([]string, []any) {
	conditions := make([]string, 0, 8)
	args := make([]any, 0, 8)
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filter.Status != nil {
		add("t.status = $%d", *filter.Status)
	}
	if filter.Priority != nil {
		add("t.priority = $%d", *filter.Priority)
	}
	if filter.CategoryID != nil {
		add("t.category_id = $%d", *filter.CategoryID)
	}
	if filter.AssignedTo != nil {
		add("t.assigned_to = $%d", *filter.AssignedTo)
	}
	if filter.CreatedBy != nil {
		add("t.created_by = $%d", *filter.CreatedBy)
	}
	if filter.CreatedFrom != nil {
		add("t.created_at >= $%d", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		add("t.created_at <= $%d", *filter.CreatedTo)
	}
	if query := strings.TrimSpace(filter.Query); query != "" {
		if id, err := uuid.Parse(query); err == nil {
			add("t.id = $%d", id)
		} else {
			add("t.title ILIKE '%%' || $%d || '%%'", query)
		}
	}
	return conditions, args
}

func ticketNotFound() error {
	return ErrTicketNotFound
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func ticketReferenceError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return ErrInvalidReference
	}
	switch pgErr.ConstraintName {
	case "tickets_category_id_fkey":
		return ErrInvalidCategory
	case "tickets_assigned_to_fkey":
		return ErrInvalidSpecialist
	default:
		return ErrInvalidReference
	}
}
