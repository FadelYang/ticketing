package attachment

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

func (r *repository) CreateBatch(ctx context.Context, items []CreateRecordInput) ([]Attachment, error) {
	created := make([]Attachment, 0, len(items))
	for _, input := range items {
		var item Attachment
		err := r.pool.QueryRow(
			ctx,
			queryCreateAttachment,
			input.Type,
			input.TargetID,
			input.UploaderID,
			input.FileName,
			input.ContentType,
			input.SizeBytes,
			input.ObjectKey,
		).Scan(
			&item.ID,
			&item.Type,
			&item.TargetID,
			&item.UploaderID,
			&item.FileName,
			&item.ContentType,
			&item.SizeBytes,
			&item.ObjectKey,
			&item.CreatedAt,
		)
		if err != nil {
			if isInvalidUUID(err) {
				return nil, fmt.Errorf("%w: invalid uuid", apperror.ErrValidation)
			}
			return nil, fmt.Errorf("create attachment: %w", err)
		}
		created = append(created, item)
	}
	return created, nil
}

func (r *repository) FindByTarget(ctx context.Context, targetType string, targetIDs []string) ([]Attachment, error) {
	rows, err := r.pool.Query(ctx, queryListAttachmentsByTargets, targetType, targetIDs)
	if err != nil {
		if isInvalidUUID(err) {
			return nil, fmt.Errorf("%w: invalid target_id uuid", apperror.ErrValidation)
		}
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	defer rows.Close()

	items := make([]Attachment, 0)
	for rows.Next() {
		var item Attachment
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.TargetID,
			&item.UploaderID,
			&item.FileName,
			&item.ContentType,
			&item.SizeBytes,
			&item.ObjectKey,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan attachment row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attachment rows: %w", err)
	}

	return items, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Attachment, error) {
	var item Attachment
	err := r.pool.QueryRow(ctx, queryGetAttachmentByID, id).Scan(
		&item.ID,
		&item.Type,
		&item.TargetID,
		&item.UploaderID,
		&item.FileName,
		&item.ContentType,
		&item.SizeBytes,
		&item.ObjectKey,
		&item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("find attachment by id: %w", err)
	}
	return &item, nil
}

func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}
