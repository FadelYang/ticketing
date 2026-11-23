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
	repo     Repository
	users    UserFinder
	comments CommentCreator
}

func NewUsecase(repo Repository, users UserFinder) Usecase {
	return &usecase{
		repo:  repo,
		users: users,
	}
}

func (u *usecase) SetCommentCreator(comments CommentCreator) {
	u.comments = comments
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
		return nil, fmt.Errorf("%w: status must be one of open, in_progress, resolved, closed", apperror.ErrValidation)
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

func (u *usecase) List(ctx context.Context, filter Filter) (*ListResult, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Status = normalizeStatus(filter.Status)
	filter.Priority = strings.ToLower(strings.TrimSpace(filter.Priority))
	filter.Category = strings.TrimSpace(filter.Category)
	filter.RequesterID = strings.TrimSpace(filter.RequesterID)
	filter.AssigneeID = strings.TrimSpace(filter.AssigneeID)
	filter.Sort = strings.ToLower(strings.TrimSpace(filter.Sort))
	if filter.Sort == "" {
		filter.Sort = SortNewest
	}
	if filter.Page == 0 {
		filter.Page = DefaultPage
	}
	if filter.Limit == 0 {
		filter.Limit = DefaultLimit
	}

	if filter.Status != "" && !IsValidStatus(filter.Status) {
		return nil, fmt.Errorf("%w: status must be one of open, in_progress, resolved, closed", apperror.ErrValidation)
	}
	if filter.Priority != "" && !IsValidPriority(filter.Priority) {
		return nil, fmt.Errorf("%w: priority must be one of low, medium, high, urgent", apperror.ErrValidation)
	}
	if filter.Sort != SortNewest && filter.Sort != SortOldest {
		return nil, fmt.Errorf("%w: sort must be newest or oldest", apperror.ErrValidation)
	}
	if filter.Page < 1 {
		return nil, fmt.Errorf("%w: page must be at least 1", apperror.ErrValidation)
	}
	if filter.Limit < 1 || filter.Limit > MaxLimit {
		return nil, fmt.Errorf("%w: limit must be between 1 and %d", apperror.ErrValidation, MaxLimit)
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
		return nil, fmt.Errorf("%w: status must be one of open, in_progress, resolved, closed", apperror.ErrValidation)
	}
	if err := u.validateParticipants(ctx, input.RequesterID, input.AssigneeID); err != nil {
		return nil, err
	}

	return u.repo.Update(ctx, id, input)
}

func (u *usecase) Patch(ctx context.Context, id string, actorID string, actorRole string, input PatchInput) (*Ticket, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}

	existing, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	merged := UpdateInput{
		Title:       existing.Title,
		Description: existing.Description,
		Category:    existing.Category,
		Priority:    existing.Priority,
		Status:      existing.Status,
		RequesterID: existing.RequesterID,
		AssigneeID:  existing.AssigneeID,
		DeadlineAt:  existing.DeadlineAt,
	}

	if input.Priority != nil {
		merged.Priority = strings.ToLower(strings.TrimSpace(*input.Priority))
	}
	if input.AssigneeID != nil {
		nextAssignee := trimOptionalString(input.AssigneeID)
		if existing.AssigneeID != nil && (nextAssignee == nil || *nextAssignee != *existing.AssigneeID) {
			if !strings.EqualFold(strings.TrimSpace(actorRole), user.RoleAdmin) {
				return nil, fmt.Errorf("%w: only admin can change assignee_id once assigned", apperror.ErrForbidden)
			}
		}
		merged.AssigneeID = nextAssignee
	}
	var statusComment *string
	if input.Status != nil {
		merged.Status = normalizeStatus(*input.Status)
		if !IsValidStatus(merged.Status) {
			return nil, fmt.Errorf("%w: status must be one of open, in_progress, resolved, closed", apperror.ErrValidation)
		}
		if err := validateStatusTransition(existing.Status, merged.Status, actorID, actorRole, existing.RequesterID, merged.AssigneeID); err != nil {
			return nil, err
		}
		isClosing := merged.Status == StatusClosed && existing.Status != StatusClosed
		isReopening := merged.Status == StatusOpen && (existing.Status == StatusResolved || existing.Status == StatusClosed)
		if isClosing || isReopening {
			statusComment = trimOptionalString(input.Comment)
			if statusComment == nil {
				if isClosing {
					return nil, fmt.Errorf("%w: comment is required when closing a ticket", apperror.ErrValidation)
				}
				return nil, fmt.Errorf("%w: comment is required when reopening a ticket", apperror.ErrValidation)
			}
		}
	}

	if err := validateTicketFields(merged.Title, merged.Description, merged.Category, merged.Priority, merged.DeadlineAt); err != nil {
		return nil, err
	}
	if err := u.validateParticipants(ctx, merged.RequesterID, merged.AssigneeID); err != nil {
		return nil, err
	}

	updated, err := u.repo.Update(ctx, id, merged)
	if err != nil {
		return nil, err
	}

	if statusComment != nil && u.comments != nil {
		if err := u.comments.CreateForTicket(ctx, actorID, actorRole, id, *statusComment); err != nil {
			return nil, err
		}
	}

	return updated, nil
}

func validateStatusTransition(fromStatus, toStatus, actorID, actorRole, requesterID string, assigneeID *string) error {
	if fromStatus == toStatus {
		return nil
	}

	role := strings.ToLower(strings.TrimSpace(actorRole))
	isSpecificAssignee := role == user.RoleAgent && assigneeID != nil && *assigneeID == actorID
	isTicketRequester := role == user.RoleRequester && requesterID == actorID

	switch fromStatus {
	case StatusOpen:
		switch toStatus {
		case StatusInProgress:
			if role == user.RoleAdmin || isSpecificAssignee {
				return nil
			}
			return fmt.Errorf("%w: only assigned agent or admin can move ticket from open to in_progress", apperror.ErrForbidden)
		case StatusClosed:
			if role == user.RoleAdmin || isTicketRequester {
				return nil
			}
			return fmt.Errorf("%w: only requester or admin can close an open ticket", apperror.ErrForbidden)
		default:
			return fmt.Errorf("%w: cannot change status from open to %s", apperror.ErrValidation, toStatus)
		}

	case StatusInProgress:
		switch toStatus {
		case StatusResolved:
			if role == user.RoleAdmin || isSpecificAssignee {
				return nil
			}
			return fmt.Errorf("%w: only assigned agent or admin can move ticket from in_progress to resolved", apperror.ErrForbidden)
		default:
			return fmt.Errorf("%w: cannot change status from in_progress to %s", apperror.ErrValidation, toStatus)
		}

	case StatusResolved:
		switch toStatus {
		case StatusClosed:
			if role == user.RoleAdmin || role == user.RoleAgent || isTicketRequester {
				return nil
			}
			return fmt.Errorf("%w: insufficient permissions to close resolved ticket", apperror.ErrForbidden)
		case StatusOpen:
			if isTicketRequester {
				return nil
			}
			return fmt.Errorf("%w: only requester can reopen a resolved ticket", apperror.ErrForbidden)
		default:
			return fmt.Errorf("%w: cannot change status from resolved to %s", apperror.ErrValidation, toStatus)
		}

	case StatusClosed:
		switch toStatus {
		case StatusOpen:
			if isTicketRequester {
				return nil
			}
			return fmt.Errorf("%w: only requester can reopen a closed ticket", apperror.ErrForbidden)
		default:
			return fmt.Errorf("%w: cannot change status from closed to %s", apperror.ErrValidation, toStatus)
		}

	default:
		return fmt.Errorf("%w: cannot change status from %s to %s", apperror.ErrValidation, fromStatus, toStatus)
	}
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
