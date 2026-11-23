package user

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

func (r *repository) Create(ctx context.Context, input CreateInput) (*User, error) {
	var item User
	err := r.pool.QueryRow(
		ctx,
		queryCreateUser,
		input.Name,
		input.Email,
		input.PhoneNumber,
		input.Role,
		input.CreatedBy,
		input.PasswordHash,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Email,
		&item.PhoneNumber,
		&item.Role,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.CreatedBy,
	)
	if err != nil {
		if mapped := mapPgWriteError(err); mapped != nil {
			return nil, mapped
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &item, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*User, error) {
	var item User
	err := r.pool.QueryRow(ctx, queryGetUserByID, id).Scan(
		&item.ID,
		&item.Name,
		&item.Email,
		&item.PhoneNumber,
		&item.Role,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.CreatedBy,
	)
	if err != nil {
		if isNotFoundOrInvalidUUID(err) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &item, nil
}

func (r *repository) FindAll(ctx context.Context, filter Filter) ([]User, error) {
	rows, err := r.pool.Query(ctx, queryListUsers, filter.Search)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	items := make([]User, 0)
	for rows.Next() {
		var item User
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Email,
			&item.PhoneNumber,
			&item.Role,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.CreatedBy,
		); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}

	return items, nil
}

func (r *repository) Update(ctx context.Context, id string, input UpdateInput) (*User, error) {
	var item User
	err := r.pool.QueryRow(
		ctx,
		queryUpdateUser,
		id,
		input.Name,
		input.Email,
		input.PhoneNumber,
		input.Role,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Email,
		&item.PhoneNumber,
		&item.Role,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.CreatedBy,
	)
	if err != nil {
		if isNotFoundOrInvalidUUID(err) {
			return nil, apperror.ErrNotFound
		}
		if mapped := mapPgWriteError(err); mapped != nil {
			return nil, mapped
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &item, nil
}

func (r *repository) UpdateRole(ctx context.Context, id string, role string) (*User, error) {
	var item User
	err := r.pool.QueryRow(ctx, queryUpdateUserRole, id, role).Scan(
		&item.ID,
		&item.Name,
		&item.Email,
		&item.PhoneNumber,
		&item.Role,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.CreatedBy,
	)
	if err != nil {
		if isNotFoundOrInvalidUUID(err) {
			return nil, apperror.ErrNotFound
		}
		if mapped := mapPgWriteError(err); mapped != nil {
			return nil, mapped
		}
		return nil, fmt.Errorf("update user role: %w", err)
	}
	return &item, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, queryDeleteUser, id)
	if err != nil {
		if isInvalidUUID(err) {
			return apperror.ErrNotFound
		}
		return fmt.Errorf("delete user: %w", err)
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
	case "23505":
		if strings.Contains(pgErr.ConstraintName, "email") {
			return fmt.Errorf("%w: email already exists", apperror.ErrConflict)
		}
		if strings.Contains(pgErr.ConstraintName, "phonenumber") {
			return fmt.Errorf("%w: phonenumber already exists", apperror.ErrConflict)
		}
		return fmt.Errorf("%w: user already exists", apperror.ErrConflict)
	case "23503":
		return fmt.Errorf("%w: created_by references a non-existent user", apperror.ErrValidation)
	default:
		return nil
	}
}
