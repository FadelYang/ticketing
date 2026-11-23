package attachment_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"monorepo/backend/internal/module/attachment"
	"monorepo/backend/internal/module/comment"
	"monorepo/backend/internal/module/ticket"
	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
)

type mockRepo struct {
	saved []attachment.Attachment
}

func (m *mockRepo) CreateBatch(_ context.Context, items []attachment.CreateRecordInput) ([]attachment.Attachment, error) {
	out := make([]attachment.Attachment, 0, len(items))
	for _, item := range items {
		att := attachment.Attachment{
			ID:          "att-1",
			Type:        item.Type,
			TargetID:    item.TargetID,
			UploaderID:  item.UploaderID,
			FileName:    item.FileName,
			ContentType: item.ContentType,
			SizeBytes:   item.SizeBytes,
			ObjectKey:   item.ObjectKey,
		}
		m.saved = append(m.saved, att)
		out = append(out, att)
	}
	return out, nil
}

func (m *mockRepo) FindByTarget(_ context.Context, _ string, _ []string) ([]attachment.Attachment, error) {
	return m.saved, nil
}

func (m *mockRepo) FindByID(_ context.Context, _ string) (*attachment.Attachment, error) {
	if len(m.saved) == 0 {
		return nil, apperror.ErrNotFound
	}
	copyAtt := m.saved[0]
	return &copyAtt, nil
}

type mockStore struct {
	uploadedKeys []string
}

func (m *mockStore) EnsureBucket(_ context.Context) error { return nil }

func (m *mockStore) Upload(_ context.Context, objectKey string, _ io.Reader, _ int64, _ string) error {
	m.uploadedKeys = append(m.uploadedKeys, objectKey)
	return nil
}

func (m *mockStore) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte("img"))), nil
}

type mockTicketFinder struct{}

func (m *mockTicketFinder) GetByID(_ context.Context, id string) (*ticket.Ticket, error) {
	return &ticket.Ticket{ID: id, RequesterID: "req-1"}, nil
}

type mockCommentFinder struct{}

func (m *mockCommentFinder) GetByID(_ context.Context, id, _, _ string) (*comment.Comment, error) {
	return &comment.Comment{ID: id, TicketID: "t-1", AuthorID: "req-1"}, nil
}

func TestAttachmentUsecase_UploadValidationAndSuccess(t *testing.T) {
	repo := &mockRepo{}
	store := &mockStore{}
	uc := attachment.NewUsecase(repo, store, &mockTicketFinder{}, &mockCommentFinder{})

	t.Run("rejects invalid type", func(t *testing.T) {
		_, err := uc.Upload(context.Background(), "req-1", user.RoleRequester, "user", "t-1", []attachment.UploadFileInput{
			{FileName: "a.png", ContentType: "image/png", SizeBytes: 10, Reader: bytes.NewReader([]byte("1234567890"))},
		})
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected validation error, got %v", err)
		}
	})

	t.Run("rejects non-image content type", func(t *testing.T) {
		_, err := uc.Upload(context.Background(), "req-1", user.RoleRequester, attachment.TypeTicket, "t-1", []attachment.UploadFileInput{
			{FileName: "doc.pdf", ContentType: "application/pdf", SizeBytes: 10, Reader: bytes.NewReader([]byte("1234567890"))},
		})
		if !errors.Is(err, apperror.ErrValidation) {
			t.Fatalf("expected validation error, got %v", err)
		}
	})

	t.Run("rejects requester uploading to another users ticket", func(t *testing.T) {
		_, err := uc.Upload(context.Background(), "req-2", user.RoleRequester, attachment.TypeTicket, "t-1", []attachment.UploadFileInput{
			{FileName: "a.png", ContentType: "image/png", SizeBytes: 10, Reader: bytes.NewReader([]byte("1234567890"))},
		})
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Fatalf("expected forbidden error, got %v", err)
		}
	})

	t.Run("uploads multiple images for ticket", func(t *testing.T) {
		items, err := uc.Upload(context.Background(), "req-1", user.RoleRequester, attachment.TypeTicket, "t-1", []attachment.UploadFileInput{
			{FileName: "one.png", ContentType: "image/png", SizeBytes: 4, Reader: bytes.NewReader([]byte("1234"))},
			{FileName: "two.jpg", ContentType: "image/jpeg", SizeBytes: 4, Reader: bytes.NewReader([]byte("5678"))},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 || len(store.uploadedKeys) != 2 {
			t.Fatalf("expected 2 uploaded attachments, got items=%d keys=%d", len(items), len(store.uploadedKeys))
		}
	})
}
