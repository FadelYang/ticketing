package example

import (
	"context"
	"errors"
	"fmt"

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

func (r *repository) Create(ctx context.Context, input CreateInput) (*Example, error) {
	var item Example
	err := r.pool.QueryRow(
		ctx,
		queryCreateExample,
		input.Name,
		input.Description,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create example: %w", err)
	}
	return &item, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Example, error) {
	var item Example
	err := r.pool.QueryRow(ctx, queryGetExampleByID, id).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if isNotFoundOrInvalidUUID(err) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("find example by id: %w", err)
	}
	return &item, nil
}

func (r *repository) FindAll(ctx context.Context, filter Filter) ([]Example, error) {
	rows, err := r.pool.Query(ctx, queryListExamples, filter.Search)
	if err != nil {
		return nil, fmt.Errorf("list examples: %w", err)
	}
	defer rows.Close()

	items := make([]Example, 0)
	for rows.Next() {
		var item Example
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan example row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate example rows: %w", err)
	}

	return items, nil
}

func (r *repository) Update(ctx context.Context, id string, input UpdateInput) (*Example, error) {
	var item Example
	err := r.pool.QueryRow(
		ctx,
		queryUpdateExample,
		id,
		input.Name,
		input.Description,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if isNotFoundOrInvalidUUID(err) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("update example: %w", err)
	}
	return &item, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, queryDeleteExample, id)
	if err != nil {
		if isInvalidUUID(err) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("delete example: %w", err)
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
	// PostgreSQL SQLSTATE 22P02 = invalid_text_representation (e.g. malformed UUID)
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}
