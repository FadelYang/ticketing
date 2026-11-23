package comment

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

func (r *repository) Create(ctx context.Context, input CreateInput) (*Comment, error) {
	var item Comment
	err := r.pool.QueryRow(
		ctx,
		queryCreateComment,
		input.TicketID,
		input.AuthorID,
		input.Body,
		input.IsInternal,
	).Scan(
		&item.ID,
		&item.TicketID,
		&item.AuthorID,
		&item.Body,
		&item.IsInternal,
		&item.CreatedAt,
	)
	if err != nil {
		if mapped := mapPgWriteError(err); mapped != nil {
			return nil, mapped
		}
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return &item, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Comment, error) {
	var item Comment
	err := r.pool.QueryRow(ctx, queryGetCommentByID, id).Scan(
		&item.ID,
		&item.TicketID,
		&item.AuthorID,
		&item.Body,
		&item.IsInternal,
		&item.CreatedAt,
	)
	if err != nil {
		if isNotFoundOrInvalidUUID(err) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("find comment by id: %w", err)
	}
	return &item, nil
}

func (r *repository) FindAll(ctx context.Context, filter Filter) (*ListResult, error) {
	var total int64
	err := r.pool.QueryRow(
		ctx,
		queryCountComments,
		filter.TicketID,
		filter.IncludeInternal,
	).Scan(&total)
	if err != nil {
		if isInvalidUUID(err) {
			return nil, fmt.Errorf("%w: invalid ticket_id uuid", apperror.ErrValidation)
		}
		return nil, fmt.Errorf("count comments: %w", err)
	}

	query := queryListCommentsOldest
	if filter.Sort == SortNewest {
		query = queryListCommentsNewest
	}
	offset := (filter.Page - 1) * filter.Limit

	rows, err := r.pool.Query(
		ctx,
		query,
		filter.TicketID,
		filter.IncludeInternal,
		filter.Limit,
		offset,
	)
	if err != nil {
		if isInvalidUUID(err) {
			return nil, fmt.Errorf("%w: invalid ticket_id uuid", apperror.ErrValidation)
		}
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	items := make([]Comment, 0)
	for rows.Next() {
		var item Comment
		if err := rows.Scan(
			&item.ID,
			&item.TicketID,
			&item.AuthorID,
			&item.Body,
			&item.IsInternal,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan comment row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comment rows: %w", err)
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
		if strings.Contains(pgErr.ConstraintName, "author") {
			return fmt.Errorf("%w: author_id references a non-existent user", apperror.ErrValidation)
		}
		return fmt.Errorf("%w: ticket_id references a non-existent ticket", apperror.ErrValidation)
	default:
		return nil
	}
}
