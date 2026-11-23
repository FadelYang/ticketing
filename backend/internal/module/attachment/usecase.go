package attachment

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
	"monorepo/backend/internal/shared/storage"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

type usecase struct {
	repo     Repository
	store    storage.ObjectStore
	tickets  TicketFinder
	comments CommentFinder
}

func NewUsecase(repo Repository, store storage.ObjectStore, tickets TicketFinder, comments CommentFinder) Usecase {
	return &usecase{
		repo:     repo,
		store:    store,
		tickets:  tickets,
		comments: comments,
	}
}

func (u *usecase) Upload(ctx context.Context, actorID, actorRole, targetType, targetID string, files []UploadFileInput) ([]Attachment, error) {
	role, err := validateActor(actorID, actorRole)
	if err != nil {
		return nil, err
	}

	targetType = strings.ToLower(strings.TrimSpace(targetType))
	targetID = strings.TrimSpace(targetID)
	if !isValidType(targetType) {
		return nil, fmt.Errorf("%w: type must be ticket or comment", apperror.ErrValidation)
	}
	if targetID == "" {
		return nil, fmt.Errorf("%w: target_id is required", apperror.ErrValidation)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("%w: at least one image file is required", apperror.ErrValidation)
	}
	if len(files) > MaxFilesPerUpload {
		return nil, fmt.Errorf("%w: maximum %d files allowed per upload", apperror.ErrValidation, MaxFilesPerUpload)
	}

	if err := u.authorizeTarget(ctx, actorID, role, targetType, targetID); err != nil {
		return nil, err
	}

	if u.store == nil {
		return nil, fmt.Errorf("object storage is not configured")
	}

	records := make([]CreateRecordInput, 0, len(files))
	for _, f := range files {
		ct := strings.ToLower(strings.TrimSpace(f.ContentType))
		if !allowedImageTypes[ct] {
			return nil, fmt.Errorf("%w: only jpeg, png, webp, and gif images are allowed", apperror.ErrValidation)
		}
		if f.SizeBytes <= 0 || f.SizeBytes > MaxFileSizeBytes {
			return nil, fmt.Errorf("%w: each image must be between 1 byte and 5 MB", apperror.ErrValidation)
		}

		safeName := filepath.Base(strings.TrimSpace(f.FileName))
		if safeName == "" || safeName == "." {
			safeName = "image"
		}
		objectKey := fmt.Sprintf("attachments/%s/%s/%s-%s", targetType, targetID, uuid.NewString(), safeName)

		if err := u.store.Upload(ctx, objectKey, f.Reader, f.SizeBytes, ct); err != nil {
			return nil, err
		}

		records = append(records, CreateRecordInput{
			Type:        targetType,
			TargetID:    targetID,
			UploaderID:  strings.TrimSpace(actorID),
			FileName:    safeName,
			ContentType: ct,
			SizeBytes:   f.SizeBytes,
			ObjectKey:   objectKey,
		})
	}

	return u.repo.CreateBatch(ctx, records)
}

func (u *usecase) ListByTarget(ctx context.Context, actorID, actorRole, targetType string, targetIDs []string) ([]Attachment, error) {
	role, err := validateActor(actorID, actorRole)
	if err != nil {
		return nil, err
	}

	targetType = strings.ToLower(strings.TrimSpace(targetType))
	if !isValidType(targetType) {
		return nil, fmt.Errorf("%w: type must be ticket or comment", apperror.ErrValidation)
	}

	cleanedIDs := make([]string, 0, len(targetIDs))
	for _, id := range targetIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			if err := u.authorizeTarget(ctx, actorID, role, targetType, trimmed); err != nil {
				return nil, err
			}
			cleanedIDs = append(cleanedIDs, trimmed)
		}
	}

	if len(cleanedIDs) == 0 {
		return []Attachment{}, nil
	}

	return u.repo.FindByTarget(ctx, targetType, cleanedIDs)
}

func (u *usecase) GetFile(ctx context.Context, id, actorID, actorRole string) (*Attachment, io.ReadCloser, error) {
	role, err := validateActor(actorID, actorRole)
	if err != nil {
		return nil, nil, err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}

	att, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if err := u.authorizeTarget(ctx, actorID, role, att.Type, att.TargetID); err != nil {
		return nil, nil, err
	}

	if u.store == nil {
		return nil, nil, fmt.Errorf("object storage is not configured")
	}

	reader, err := u.store.Get(ctx, att.ObjectKey)
	if err != nil {
		return nil, nil, err
	}

	return att, reader, nil
}

func (u *usecase) authorizeTarget(ctx context.Context, actorID, role, targetType, targetID string) error {
	switch targetType {
	case TypeTicket:
		t, err := u.tickets.GetByID(ctx, targetID)
		if err != nil {
			return err
		}
		if role == user.RoleRequester && t.RequesterID != strings.TrimSpace(actorID) {
			return fmt.Errorf("%w: you can only access attachments on your own tickets", apperror.ErrForbidden)
		}
		return nil
	case TypeComment:
		_, err := u.comments.GetByID(ctx, targetID, actorID, role)
		return err
	default:
		return fmt.Errorf("%w: type must be ticket or comment", apperror.ErrValidation)
	}
}

func validateActor(actorID, actorRole string) (string, error) {
	if strings.TrimSpace(actorID) == "" {
		return "", fmt.Errorf("%w: missing authenticated user", apperror.ErrUnauthorized)
	}
	role := strings.ToLower(strings.TrimSpace(actorRole))
	if !user.IsValidRole(role) {
		return "", fmt.Errorf("%w: insufficient permissions", apperror.ErrForbidden)
	}
	return role, nil
}

func isValidType(targetType string) bool {
	return targetType == TypeTicket || targetType == TypeComment
}
