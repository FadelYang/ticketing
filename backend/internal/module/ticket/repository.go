package ticket

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"monorepo/backend/internal/shared/apperror"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, input CreateInput) (*Ticket, error) {
	var item Ticket
	err := r.pool.QueryRow(
		ctx,
		queryCreateTicket,
		input.Title,
		input.Description,
		input.Category,
		input.Priority,
		input.Status,
		input.RequesterID,
		input.AssigneeID,
		input.DeadlineAt,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Category,
		&item.Priority,
		&item.Status,
		&item.RequesterID,
		&item.AssigneeID,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeadlineAt,
		&item.IsLate,
	)
	if err != nil {
		if mapped := mapPgWriteError(err); mapped != nil {
			return nil, mapped
		}
		return nil, fmt.Errorf("create ticket: %w", err)
	}
	return &item, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Ticket, error) {
	var item Ticket
	err := r.pool.QueryRow(ctx, queryGetTicketByID, id).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Category,
		&item.Priority,
		&item.Status,
		&item.RequesterID,
		&item.AssigneeID,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeadlineAt,
		&item.IsLate,
	)
	if err != nil {
		if isNotFoundOrInvalidUUID(err) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("find ticket by id: %w", err)
	}
	return &item, nil
}

func (r *repository) FindAll(ctx context.Context, filter Filter) (*ListResult, error) {
	var total int64
	err := r.pool.QueryRow(
		ctx,
		queryCountTickets,
		filter.Search,
		filter.Status,
		filter.Priority,
		filter.Category,
		filter.RequesterID,
		filter.AssigneeID,
	).Scan(&total)
	if err != nil {
		if isInvalidUUID(err) {
			return nil, fmt.Errorf("%w: invalid uuid filter", apperror.ErrValidation)
		}
		return nil, fmt.Errorf("count tickets: %w", err)
	}

	query := queryListTicketsNewest
	if filter.Sort == SortOldest {
		query = queryListTicketsOldest
	}
	offset := (filter.Page - 1) * filter.Limit

	rows, err := r.pool.Query(
		ctx,
		query,
		filter.Search,
		filter.Status,
		filter.Priority,
		filter.Category,
		filter.RequesterID,
		filter.AssigneeID,
		filter.Limit,
		offset,
	)
	if err != nil {
		if isInvalidUUID(err) {
			return nil, fmt.Errorf("%w: invalid uuid filter", apperror.ErrValidation)
		}
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	items := make([]Ticket, 0)
	for rows.Next() {
		var item Ticket
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Category,
			&item.Priority,
			&item.Status,
			&item.RequesterID,
			&item.AssigneeID,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeadlineAt,
			&item.IsLate,
		); err != nil {
			return nil, fmt.Errorf("scan ticket row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket rows: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(filter.Limit) - 1) / int64(filter.Limit))
	}

	return &ListResult{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (r *repository) FindLateUrgent(ctx context.Context, requesterID string) ([]Ticket, error) {
	rows, err := r.pool.Query(ctx, queryListLateUrgentTickets, requesterID)
	if err != nil {
		if isInvalidUUID(err) {
			return nil, fmt.Errorf("%w: invalid requester_id uuid", apperror.ErrValidation)
		}
		return nil, fmt.Errorf("list late urgent tickets: %w", err)
	}
	defer rows.Close()

	items := make([]Ticket, 0)
	for rows.Next() {
		var item Ticket
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Category,
			&item.Priority,
			&item.Status,
			&item.RequesterID,
			&item.AssigneeID,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeadlineAt,
			&item.IsLate,
		); err != nil {
			return nil, fmt.Errorf("scan late urgent ticket row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate late urgent ticket rows: %w", err)
	}

	return items, nil
}

func (r *repository) CountByStatus(ctx context.Context, filter Filter) (*StatusCounts, error) {
	rows, err := r.pool.Query(
		ctx,
		queryCountTicketsByStatus,
		filter.Search,
		filter.Priority,
		filter.Category,
		filter.RequesterID,
		filter.AssigneeID,
	)
	if err != nil {
		if isInvalidUUID(err) {
			return nil, fmt.Errorf("%w: invalid uuid filter", apperror.ErrValidation)
		}
		return nil, fmt.Errorf("count tickets by status: %w", err)
	}
	defer rows.Close()

	counts := &StatusCounts{}
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan ticket status count: %w", err)
		}
		switch status {
		case StatusOpen:
			counts.Open = count
		case StatusInProgress:
			counts.InProgress = count
		case StatusResolved:
			counts.Resolved = count
		case StatusClosed:
			counts.Closed = count
		}
		counts.All += count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket status count rows: %w", err)
	}

	return counts, nil
}

func (r *repository) Update(ctx context.Context, id string, input UpdateInput) (*Ticket, error) {
	var item Ticket
	err := r.pool.QueryRow(
		ctx,
		queryUpdateTicket,
		id,
		input.Title,
		input.Description,
		input.Category,
		input.Priority,
		input.Status,
		input.RequesterID,
		input.AssigneeID,
		input.DeadlineAt,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Category,
		&item.Priority,
		&item.Status,
		&item.RequesterID,
		&item.AssigneeID,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeadlineAt,
		&item.IsLate,
	)
	if err != nil {
		if isNotFoundOrInvalidUUID(err) {
			return nil, apperror.ErrNotFound
		}
		if mapped := mapPgWriteError(err); mapped != nil {
			return nil, mapped
		}
		return nil, fmt.Errorf("update ticket: %w", err)
	}
	return &item, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, queryDeleteTicket, id)
	if err != nil {
		if isInvalidUUID(err) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("delete ticket: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func isNotFoundOrInvalidUUID(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err)
}

func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}

func mapPgWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	switch pgErr.Code {
	case "22P02":
		return fmt.Errorf("%w: invalid uuid format", apperror.ErrValidation)
	case "23503":
		if strings.Contains(pgErr.ConstraintName, "assignee") {
			return fmt.Errorf("%w: assignee_id references a non-existent user", apperror.ErrValidation)
		}
		return fmt.Errorf("%w: requester_id references a non-existent user", apperror.ErrValidation)
	case "23514":
		return fmt.Errorf("%w: invalid field value", apperror.ErrValidation)
	default:
		return nil
	}
}
