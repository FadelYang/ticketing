package comment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"monorepo/backend/internal/module/ticket"
	"monorepo/backend/internal/module/user"
	"monorepo/backend/internal/shared/apperror"
)

type usecase struct {
	repo    Repository
	tickets TicketFinder
	users   UserFinder
}

func NewUsecase(repo Repository, tickets TicketFinder, users UserFinder) Usecase {
	return &usecase{
		repo:    repo,
		tickets: tickets,
		users:   users,
	}
}

func (u *usecase) Create(ctx context.Context, actorID string, actorRole string, input CreateInput) (*Comment, error) {
	role, err := validateActor(actorID, actorRole)
	if err != nil {
		return nil, err
	}

	input.TicketID = strings.TrimSpace(input.TicketID)
	input.AuthorID = strings.TrimSpace(actorID)
	input.Body = strings.TrimSpace(input.Body)

	if input.TicketID == "" {
		return nil, fmt.Errorf("%w: ticket_id is required", apperror.ErrValidation)
	}
	if input.Body == "" {
		return nil, fmt.Errorf("%w: body is required", apperror.ErrValidation)
	}
	if input.IsInternal && role == user.RoleRequester {
		return nil, fmt.Errorf("%w: only agent or admin can create internal comments", apperror.ErrForbidden)
	}

	t, err := u.tickets.GetByID(ctx, input.TicketID)
	if err != nil {
		return nil, err
	}
	if err := authorizeTicketAccess(t, input.AuthorID, role); err != nil {
		return nil, err
	}

	if _, err := u.users.GetByID(ctx, input.AuthorID); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, fmt.Errorf("%w: author_id references a non-existent user", apperror.ErrValidation)
		}
		return nil, err
	}

	return u.repo.Create(ctx, input)
}

func (u *usecase) GetByID(ctx context.Context, id string, actorID string, actorRole string) (*Comment, error) {
	role, err := validateActor(actorID, actorRole)
	if err != nil {
		return nil, err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", apperror.ErrValidation)
	}

	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role == user.RoleRequester {
		if item.IsInternal {
			return nil, fmt.Errorf("%w: cannot access internal comments", apperror.ErrForbidden)
		}
		t, err := u.tickets.GetByID(ctx, item.TicketID)
		if err != nil {
			return nil, err
		}
		if err := authorizeTicketAccess(t, strings.TrimSpace(actorID), role); err != nil {
			return nil, err
		}
	}

	return item, nil
}

func (u *usecase) ListByTicket(ctx context.Context, actorID string, actorRole string, filter Filter) (*ListResult, error) {
	role, err := validateActor(actorID, actorRole)
	if err != nil {
		return nil, err
	}

	filter.TicketID = strings.TrimSpace(filter.TicketID)
	if filter.TicketID == "" {
		return nil, fmt.Errorf("%w: ticket_id is required", apperror.ErrValidation)
	}

	filter.Sort = strings.ToLower(strings.TrimSpace(filter.Sort))
	if filter.Sort == "" {
		filter.Sort = SortOldest
	}
	if filter.Sort != SortNewest && filter.Sort != SortOldest {
		return nil, fmt.Errorf("%w: sort must be newest or oldest", apperror.ErrValidation)
	}
	if filter.Page == 0 {
		filter.Page = DefaultPage
	}
	if filter.Page < 1 {
		return nil, fmt.Errorf("%w: page must be at least 1", apperror.ErrValidation)
	}
	if filter.Limit == 0 {
		filter.Limit = DefaultLimit
	}
	if filter.Limit < 1 || filter.Limit > MaxLimit {
		return nil, fmt.Errorf("%w: limit must be between 1 and %d", apperror.ErrValidation, MaxLimit)
	}

	t, err := u.tickets.GetByID(ctx, filter.TicketID)
	if err != nil {
		return nil, err
	}
	if err := authorizeTicketAccess(t, strings.TrimSpace(actorID), role); err != nil {
		return nil, err
	}

	filter.IncludeInternal = role == user.RoleAgent || role == user.RoleAdmin

	return u.repo.FindAll(ctx, filter)
}

func validateActor(actorID string, actorRole string) (string, error) {
	if strings.TrimSpace(actorID) == "" {
		return "", fmt.Errorf("%w: missing authenticated user", apperror.ErrUnauthorized)
	}
	role := strings.ToLower(strings.TrimSpace(actorRole))
	if !user.IsValidRole(role) {
		return "", fmt.Errorf("%w: insufficient permissions", apperror.ErrForbidden)
	}
	return role, nil
}

func authorizeTicketAccess(t *ticket.Ticket, actorID string, role string) error {
	if role == user.RoleRequester && t.RequesterID != actorID {
		return fmt.Errorf("%w: you can only access comments on your own tickets", apperror.ErrForbidden)
	}
	return nil
}
