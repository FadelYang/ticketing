package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type SeedResult struct {
	ExistingRequesters int
	ExistingAgents     int
	ExistingAdmins     int
	UsersInserted      int
	TicketsInserted    int
	CommentsInserted   int
}

type roleUsers struct {
	requesters []string
	agents     []string
	admins     []string
}

type seedTicket struct {
	ID             string
	Title          string
	Description    string
	Category       string
	Priority       string
	Status         string
	RequesterID    string
	AssigneeID     string
	DeadlineOffset time.Duration
}

type seedComment struct {
	ID         string
	TicketID   string
	AuthorID   string
	Body       string
	IsInternal bool
	AgeOffset  time.Duration
}

// Seed populates sample tickets and comments using existing users in the database
// (creating fallback users only if a required role has no existing rows).
func Seed(ctx context.Context, pool *pgxpool.Pool) (*SeedResult, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin seed transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	usersByRole, usersInserted, err := loadOrSeedUsers(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("resolve users: %w", err)
	}

	ticketsInserted, err := seedTickets(ctx, tx, usersByRole)
	if err != nil {
		return nil, fmt.Errorf("seed tickets: %w", err)
	}

	commentsInserted, err := seedComments(ctx, tx, usersByRole)
	if err != nil {
		return nil, fmt.Errorf("seed comments: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit seed transaction: %w", err)
	}

	return &SeedResult{
		ExistingRequesters: len(usersByRole.requesters),
		ExistingAgents:     len(usersByRole.agents),
		ExistingAdmins:     len(usersByRole.admins),
		UsersInserted:      usersInserted,
		TicketsInserted:    ticketsInserted,
		CommentsInserted:   commentsInserted,
	}, nil
}

func loadOrSeedUsers(ctx context.Context, tx pgx.Tx) (roleUsers, int, error) {
	var result roleUsers

	rows, err := tx.Query(ctx, `SELECT id, role FROM users ORDER BY created_at ASC`)
	if err != nil {
		return result, 0, fmt.Errorf("query existing users: %w", err)
	}
	for rows.Next() {
		var id, role string
		if err := rows.Scan(&id, &role); err != nil {
			rows.Close()
			return result, 0, fmt.Errorf("scan user: %w", err)
		}
		switch role {
		case "requester":
			result.requesters = append(result.requesters, id)
		case "agent":
			result.agents = append(result.agents, id)
		case "admin":
			result.admins = append(result.admins, id)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return result, 0, fmt.Errorf("iterate users: %w", err)
	}

	inserted := 0
	defaultUsers := []struct {
		name     string
		email    string
		phone    string
		role     string
		password string
	}{
		{
			name:     "Admin",
			email:    "admin@gmail.com",
			phone:    "+628000000001",
			role:     "admin",
			password: "Ticketing@123",
		},
		{
			name:     "Agent",
			email:    "agent@gamil.com",
			phone:    "+628000000002",
			role:     "agent",
			password: "Ticketing@123",
		},
		{
			name:     "Requester",
			email:    "requester@gamil.com",
			phone:    "+628000000003",
			role:     "requester",
			password: "Ticketing@123",
		},
	}

	for _, u := range defaultUsers {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			return result, 0, fmt.Errorf("hash password: %w", err)
		}
		tag, err := tx.Exec(
			ctx,
			`INSERT INTO users (name, email, phonenumber, role, password_hash)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (email) DO NOTHING`,
			u.name, u.email, u.phone, u.role, string(hash),
		)
		if err != nil {
			return result, 0, fmt.Errorf("insert default %s: %w", u.role, err)
		}
		if tag.RowsAffected() > 0 {
			inserted++
			var id string
			if err := tx.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, u.email).Scan(&id); err != nil {
				return result, 0, fmt.Errorf("lookup default %s: %w", u.role, err)
			}
			switch u.role {
			case "requester":
				result.requesters = append(result.requesters, id)
			case "agent":
				result.agents = append(result.agents, id)
			case "admin":
				result.admins = append(result.admins, id)
			}
		}
	}

	return result, inserted, nil
}

func pickID(ids []string, index int) string {
	return ids[index%len(ids)]
}

func seedTickets(ctx context.Context, tx pgx.Tx, users roleUsers) (int, error) {
	now := time.Now().UTC()

	tickets := []seedTicket{
		{
			ID:             "11111111-1111-4111-8111-111111111101",
			Title:          "Kafka stream disconnects during ingest spikes",
			Description:    "Consumer offset drift exceeds 45,000 records on cluster-04 during peak traffic. Immediate failover investigation requested.",
			Category:       "Infrastructure",
			Priority:       "urgent",
			Status:         "open",
			RequesterID:    pickID(users.requesters, 0),
			AssigneeID:     pickID(users.agents, 0),
			DeadlineOffset: 4 * time.Hour,
		},
		{
			ID:             "11111111-1111-4111-8111-111111111102",
			Title:          "Payment webhook timeout on Stripe v3 integration",
			Description:    "Webhook retries are exceeding the exponential backoff budget, causing delayed invoice status updates on the merchant dashboard.",
			Category:       "Billing & Payments",
			Priority:       "high",
			Status:         "in_progress",
			RequesterID:    pickID(users.requesters, 0),
			AssigneeID:     pickID(users.agents, 1),
			DeadlineOffset: 8 * time.Hour,
		},
		{
			ID:             "11111111-1111-4111-8111-111111111103",
			Title:          "SSO authentication SAML assertion format error",
			Description:    "Okta IDP integration failing for tenant enterprise_eu_77 due to missing NameID format in SAML response.",
			Category:       "Authentication",
			Priority:       "medium",
			Status:         "open",
			RequesterID:    pickID(users.requesters, 1),
			AssigneeID:     "",
			DeadlineOffset: 24 * time.Hour,
		},
		{
			ID:             "11111111-1111-4111-8111-111111111104",
			Title:          "PostgreSQL read replica replication lag > 120s",
			Description:    "Secondary database instance experiencing sustained write queue bloat during nightly analytics batch jobs.",
			Category:       "Database",
			Priority:       "high",
			Status:         "in_progress",
			RequesterID:    pickID(users.requesters, 1),
			AssigneeID:     pickID(users.agents, 0),
			DeadlineOffset: 12 * time.Hour,
		},
		{
			ID:             "11111111-1111-4111-8111-111111111105",
			Title:          "API Gateway throttling legitimate enterprise tier requests",
			Description:    "Token bucket rate limiter is resetting TTL windows prematurely across regional edge nodes.",
			Category:       "API & Gateway",
			Priority:       "urgent",
			Status:         "open",
			RequesterID:    pickID(users.requesters, 0),
			AssigneeID:     "",
			DeadlineOffset: 6 * time.Hour,
		},
		{
			ID:             "11111111-1111-4111-8111-111111111106",
			Title:          "Export CSV broken for date ranges over 90 days",
			Description:    "Async export worker runs out of memory when streaming more than 100k rows without cursor pagination.",
			Category:       "Backend API",
			Priority:       "low",
			Status:         "resolved",
			RequesterID:    pickID(users.requesters, 0),
			AssigneeID:     pickID(users.agents, 1),
			DeadlineOffset: 48 * time.Hour,
		},
		{
			ID:             "11111111-1111-4111-8111-111111111107",
			Title:          "Mobile navigation drawer clipping on iOS Safari",
			Description:    "Sidebar drawer bottom footer is clipped behind the browser toolbar on iOS 17 devices.",
			Category:       "Frontend UI",
			Priority:       "medium",
			Status:         "resolved",
			RequesterID:    pickID(users.requesters, 1),
			AssigneeID:     pickID(users.agents, 0),
			DeadlineOffset: 72 * time.Hour,
		},
		{
			ID:             "11111111-1111-4111-8111-111111111108",
			Title:          "Update TLS cipher suite on legacy webhook endpoint",
			Description:    "Deprecated TLS 1.1 ciphers removed from edge load balancer configuration.",
			Category:       "Infrastructure",
			Priority:       "low",
			Status:         "closed",
			RequesterID:    pickID(users.requesters, 0),
			AssigneeID:     pickID(users.agents, 0),
			DeadlineOffset: -24 * time.Hour,
		},
	}

	inserted := 0
	for _, t := range tickets {
		var assigneeID *string
		if t.AssigneeID != "" {
			id := t.AssigneeID
			assigneeID = &id
		}

		tag, err := tx.Exec(
			ctx,
			`INSERT INTO tickets (id, title, description, category, priority, status, requester_id, assignee_id, deadline_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (id) DO NOTHING`,
			t.ID, t.Title, t.Description, t.Category, t.Priority, t.Status, t.RequesterID, assigneeID, now.Add(t.DeadlineOffset),
		)
		if err != nil {
			return 0, fmt.Errorf("insert ticket %s: %w", t.ID, err)
		}
		if tag.RowsAffected() > 0 {
			inserted++
		}
	}

	return inserted, nil
}

func seedComments(ctx context.Context, tx pgx.Tx, users roleUsers) (int, error) {
	now := time.Now().UTC()

	comments := []seedComment{
		{
			ID:         "22222222-2222-4222-8222-222222222201",
			TicketID:   "11111111-1111-4111-8111-111111111101",
			AuthorID:   pickID(users.requesters, 0),
			Body:       "Attached broker metrics show consumer group rebalance every 90 seconds since 14:00 UTC.",
			IsInternal: false,
			AgeOffset:  -3 * time.Hour,
		},
		{
			ID:         "22222222-2222-4222-8222-222222222202",
			TicketID:   "11111111-1111-4111-8111-111111111101",
			AuthorID:   pickID(users.agents, 0),
			Body:       "Internal note: Checking max.poll.interval.ms and partition skew on cluster-04.",
			IsInternal: true,
			AgeOffset:  -2 * time.Hour,
		},
		{
			ID:         "22222222-2222-4222-8222-222222222203",
			TicketID:   "11111111-1111-4111-8111-111111111102",
			AuthorID:   pickID(users.agents, 1),
			Body:       "We reproduced the timeout in staging and are increasing the webhook worker concurrency.",
			IsInternal: false,
			AgeOffset:  -5 * time.Hour,
		},
		{
			ID:         "22222222-2222-4222-8222-222222222204",
			TicketID:   "11111111-1111-4111-8111-111111111102",
			AuthorID:   pickID(users.requesters, 0),
			Body:       "Thanks! Let us know once the fix is rolled out to production.",
			IsInternal: false,
			AgeOffset:  -4 * time.Hour,
		},
		{
			ID:         "22222222-2222-4222-8222-222222222205",
			TicketID:   "11111111-1111-4111-8111-111111111104",
			AuthorID:   pickID(users.agents, 0),
			Body:       "Internal note: Read replica IOPS maxed out; coordinating with DBA to throttle analytics batch.",
			IsInternal: true,
			AgeOffset:  -6 * time.Hour,
		},
		{
			ID:         "22222222-2222-4222-8222-222222222206",
			TicketID:   "11111111-1111-4111-8111-111111111106",
			AuthorID:   pickID(users.agents, 1),
			Body:       "Deployed cursor-based streaming for CSV exports. Please verify with a 120-day range.",
			IsInternal: false,
			AgeOffset:  -10 * time.Hour,
		},
		{
			ID:         "22222222-2222-4222-8222-222222222207",
			TicketID:   "11111111-1111-4111-8111-111111111107",
			AuthorID:   pickID(users.agents, 0),
			Body:       "Updated root layout containers to use dynamic viewport height (dvh).",
			IsInternal: false,
			AgeOffset:  -12 * time.Hour,
		},
		{
			ID:         "22222222-2222-4222-8222-222222222208",
			TicketID:   "11111111-1111-4111-8111-111111111108",
			AuthorID:   pickID(users.requesters, 0),
			Body:       "Confirmed TLS 1.2/1.3 handshake succeeds cleanly. Closing this ticket.",
			IsInternal: false,
			AgeOffset:  -26 * time.Hour,
		},
	}

	inserted := 0
	for _, c := range comments {
		tag, err := tx.Exec(
			ctx,
			`INSERT INTO comments (id, ticket_id, author_id, body, is_internal, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (id) DO NOTHING`,
			c.ID, c.TicketID, c.AuthorID, c.Body, c.IsInternal, now.Add(c.AgeOffset),
		)
		if err != nil {
			return 0, fmt.Errorf("insert comment %s: %w", c.ID, err)
		}
		if tag.RowsAffected() > 0 {
			inserted++
		}
	}

	return inserted, nil
}
