package ticket

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
)

type usecase struct {
	repo  Repository
	users UserFinder
}

func NewUsecase(repo Repository, users UserFinder) Usecase {
	return &usecase{
		repo:  repo,
		users: users,
	}
}

func (u *usecase) Create(ctx context.Context, input CreateInput) (*Ticket, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Category = strings.TrimSpace(input.Category)
	input.Priority = strings.ToLower(strings.TrimSpace(input.Priority))
	input.Status = normalizeStatus(input.Status)
	if input.Status == "" {
		input.Status = StatusOpen
	}
	input.RequesterID = strings.TrimSpace(input.RequesterID)
	input.AssigneeID = trimOptionalString(input.AssigneeID)

	if err := validateTicketFields(input.Title, input.Description, input.Category, input.Priority, input.DeadlineAt); err != nil {
		return nil, err
	}
	if !IsValidStatus(input.Status) {
		return nil, fmt.Errorf("%w: status must be one of open, in_progress, resolved", apperror.ErrValidation)
	}
	if err := u.validateParticipants(ctx, input.RequesterID, input.AssigneeID); err != nil {
		return nil, err
	}

	return u.repo.Create(ctx, input)
}

func (u *usecase) GetByID(ctx context.Context, id string) (*Ticket, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}
	return u.repo.FindByID(ctx, id)
}

func (u *usecase) List(ctx context.Context, filter Filter) ([]Ticket, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Status = normalizeStatus(filter.Status)
	filter.Priority = strings.ToLower(strings.TrimSpace(filter.Priority))
	filter.Category = strings.TrimSpace(filter.Category)
	filter.RequesterID = strings.TrimSpace(filter.RequesterID)
	filter.AssigneeID = strings.TrimSpace(filter.AssigneeID)

	if filter.Status != "" && !IsValidStatus(filter.Status) {
		return nil, fmt.Errorf("%w: status must be one of open, in_progress, resolved", apperror.ErrValidation)
	}
	if filter.Priority != "" && !IsValidPriority(filter.Priority) {
		return nil, fmt.Errorf("%w: priority must be one of low, medium, high, urgent", apperror.ErrValidation)
	}

	return u.repo.FindAll(ctx, filter)
}

func (u *usecase) Update(ctx context.Context, id string, input UpdateInput) (*Ticket, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Category = strings.TrimSpace(input.Category)
	input.Priority = strings.ToLower(strings.TrimSpace(input.Priority))
	input.Status = normalizeStatus(input.Status)
	input.RequesterID = strings.TrimSpace(input.RequesterID)
	input.AssigneeID = trimOptionalString(input.AssigneeID)

	if err := validateTicketFields(input.Title, input.Description, input.Category, input.Priority, input.DeadlineAt); err != nil {
		return nil, err
	}
	if input.Status != "" && !IsValidStatus(input.Status) {
		return nil, fmt.Errorf("%w: status must be one of open, in_progress, resolved", apperror.ErrValidation)
	}
	if err := u.validateParticipants(ctx, input.RequesterID, input.AssigneeID); err != nil {
		return nil, err
	}

	return u.repo.Update(ctx, id, input)
}

func (u *usecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}
	return u.repo.Delete(ctx, id)
}

func (u *usecase) validateParticipants(ctx context.Context, requesterID string, assigneeID *string) error {
	if requesterID == "" {
		return fmt.Errorf("%w: requester_id is required", apperror.ErrValidation)
	}

	requester, err := u.users.GetByID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return fmt.Errorf("%w: requester_id references a non-existent user", apperror.ErrValidation)
		}
		return err
	}
	if requester.Role != user.RoleRequester {
		return fmt.Errorf("%w: requester_id must reference a user with role requester", apperror.ErrValidation)
	}

	if assigneeID != nil {
		assignee, err := u.users.GetByID(ctx, *assigneeID)
		if err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
				return fmt.Errorf("%w: assignee_id references a non-existent user", apperror.ErrValidation)
			}
			return err
		}
		if assignee.Role != user.RoleAgent && assignee.Role != user.RoleAdmin {
			return fmt.Errorf("%w: assignee_id must reference a user with role agent or admin", apperror.ErrValidation)
		}
	}

	return nil
}

func validateTicketFields(title, description, category, priority string, deadlineAt time.Time) error {
	if title == "" {
		return fmt.Errorf("%w: title is required", apperror.ErrValidation)
	}
	if len(title) > 250 {
		return fmt.Errorf("%w: title must be at most 250 characters", apperror.ErrValidation)
	}
	if description == "" {
		return fmt.Errorf("%w: description is required", apperror.ErrValidation)
	}
	if category == "" {
		return fmt.Errorf("%w: category is required", apperror.ErrValidation)
	}
	if len(category) > 100 {
		return fmt.Errorf("%w: category must be at most 100 characters", apperror.ErrValidation)
	}
	if !IsValidPriority(priority) {
		return fmt.Errorf("%w: priority must be one of low, medium, high, urgent", apperror.ErrValidation)
	}
	if deadlineAt.IsZero() {
		return fmt.Errorf("%w: deadline_at is required", apperror.ErrValidation)
	}
	return nil
}

func normalizeStatus(status string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "in progress" {
		return StatusInProgress
	}
	return s
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
