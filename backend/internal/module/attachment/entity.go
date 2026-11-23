package attachment

import (
	"context"
	"io"
	"time"

	"monorepo/backend/internal/module/comment"
	"monorepo/backend/internal/module/ticket"
)

const (
	TypeTicket  = "ticket"
	TypeComment = "comment"

	MaxFileSizeBytes  = 5 * 1024 * 1024 // 5 MB
	MaxFilesPerUpload = 5
)

type Attachment struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	TargetID    string    `json:"target_id"`
	UploaderID  string    `json:"uploader_id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	ObjectKey   string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

type UploadFileInput struct {
	FileName    string
	ContentType string
	SizeBytes   int64
	Reader      io.Reader
}

type CreateRecordInput struct {
	Type        string
	TargetID    string
	UploaderID  string
	FileName    string
	ContentType string
	SizeBytes   int64
	ObjectKey   string
}

type TicketFinder interface {
	GetByID(ctx context.Context, id string) (*ticket.Ticket, error)
}

type CommentFinder interface {
	GetByID(ctx context.Context, id string, actorID string, actorRole string) (*comment.Comment, error)
}

type Repository interface {
	CreateBatch(ctx context.Context, items []CreateRecordInput) ([]Attachment, error)
	FindByTarget(ctx context.Context, targetType string, targetIDs []string) ([]Attachment, error)
	FindByID(ctx context.Context, id string) (*Attachment, error)
}

type Usecase interface {
	Upload(ctx context.Context, actorID, actorRole, targetType, targetID string, files []UploadFileInput) ([]Attachment, error)
	ListByTarget(ctx context.Context, actorID, actorRole, targetType string, targetIDs []string) ([]Attachment, error)
	GetFile(ctx context.Context, id, actorID, actorRole string) (*Attachment, io.ReadCloser, error)
}
