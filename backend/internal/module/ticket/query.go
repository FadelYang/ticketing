package ticket

const (
	queryCreateTicket = `
		INSERT INTO tickets (title, description, category, priority, status, requester_id, assignee_id, deadline_at)
		VALUES ($1, $2, $3, $4, $5, $6::uuid, $7::uuid, $8)
		RETURNING id::text, title, description, category, priority, status, requester_id::text, assignee_id::text, created_at, updated_at, deadline_at
	`

	queryGetTicketByID = `
		SELECT id::text, title, description, category, priority, status, requester_id::text, assignee_id::text, created_at, updated_at, deadline_at
		FROM tickets
		WHERE id = $1::uuid
	`

	queryCountTickets = `
		SELECT COUNT(*)
		FROM tickets
		WHERE ($1 = '' OR title ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR priority = $3)
		  AND ($4 = '' OR category = $4)
		  AND ($5 = '' OR requester_id = NULLIF($5, '')::uuid)
		  AND ($6 = '' OR assignee_id = NULLIF($6, '')::uuid)
	`

	queryCountTicketsByStatus = `
		SELECT status, COUNT(*)
		FROM tickets
		WHERE ($1 = '' OR title ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR priority = $2)
		  AND ($3 = '' OR category = $3)
		  AND ($4 = '' OR requester_id = NULLIF($4, '')::uuid)
		  AND ($5 = '' OR assignee_id = NULLIF($5, '')::uuid)
		GROUP BY status
	`

	queryListTicketsNewest = `
		SELECT id::text, title, description, category, priority, status, requester_id::text, assignee_id::text, created_at, updated_at, deadline_at
		FROM tickets
		WHERE ($1 = '' OR title ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR priority = $3)
		  AND ($4 = '' OR category = $4)
		  AND ($5 = '' OR requester_id = NULLIF($5, '')::uuid)
		  AND ($6 = '' OR assignee_id = NULLIF($6, '')::uuid)
		ORDER BY created_at DESC
		LIMIT $7 OFFSET $8
	`

	queryListTicketsOldest = `
		SELECT id::text, title, description, category, priority, status, requester_id::text, assignee_id::text, created_at, updated_at, deadline_at
		FROM tickets
		WHERE ($1 = '' OR title ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR priority = $3)
		  AND ($4 = '' OR category = $4)
		  AND ($5 = '' OR requester_id = NULLIF($5, '')::uuid)
		  AND ($6 = '' OR assignee_id = NULLIF($6, '')::uuid)
		ORDER BY created_at ASC
		LIMIT $7 OFFSET $8
	`

	queryUpdateTicket = `
		UPDATE tickets
		SET title        = $2,
		    description  = $3,
		    category     = $4,
		    priority     = $5,
		    status       = COALESCE(NULLIF($6, ''), status),
		    requester_id = $7::uuid,
		    assignee_id  = $8::uuid,
		    deadline_at  = $9,
		    updated_at   = NOW()
		WHERE id = $1::uuid
		RETURNING id::text, title, description, category, priority, status, requester_id::text, assignee_id::text, created_at, updated_at, deadline_at
	`

	queryDeleteTicket = `
		DELETE FROM tickets
		WHERE id = $1::uuid
	`
)
