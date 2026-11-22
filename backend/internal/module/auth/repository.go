package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreateUser(ctx context.Context, name, email, phoneNumber, passwordHash string) (*user.User, error) {
	var item user.User
	err := r.pool.QueryRow(
		ctx,
		queryRegisterUser,
		name,
		email,
		phoneNumber,
		passwordHash,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Email,
		&item.PhoneNumber,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.CreatedBy,
	)
	if err != nil {
		if mapped := mapPgWriteError(err); mapped != nil {
			return nil, mapped
		}
		return nil, fmt.Errorf("register user: %w", err)
	}
	return &item, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*userWithPassword, error) {
	var result userWithPassword
	err := r.pool.QueryRow(ctx, queryFindUserByEmail, email).Scan(
		&result.User.ID,
		&result.User.Name,
		&result.User.Email,
		&result.User.PhoneNumber,
		&result.PasswordHash,
		&result.User.CreatedAt,
		&result.User.UpdatedAt,
		&result.User.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &result, nil
}

func mapPgWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	if pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "email") {
			return fmt.Errorf("%w: email already exists", apperror.ErrConflict)
		}
		if strings.Contains(pgErr.ConstraintName, "phonenumber") {
			return fmt.Errorf("%w: phonenumber already exists", apperror.ErrConflict)
		}
		return fmt.Errorf("%w: user already exists", apperror.ErrConflict)
	}

	return nil
}
