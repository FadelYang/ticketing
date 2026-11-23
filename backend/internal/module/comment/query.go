package comment

const (
	queryCreateComment = `
		INSERT INTO comments (ticket_id, author_id, body, is_internal)
		VALUES ($1::uuid, $2::uuid, $3, $4)
		RETURNING id::text, ticket_id::text, author_id::text, body, is_internal, created_at
	`

	queryGetCommentByID = `
		SELECT id::text, ticket_id::text, author_id::text, body, is_internal, created_at
		FROM comments
		WHERE id = $1::uuid
	`

	queryCountComments = `
		SELECT COUNT(*)
		FROM comments
		WHERE ticket_id = $1::uuid
		  AND ($2::boolean OR is_internal = FALSE)
	`

	queryListCommentsOldest = `
		SELECT id::text, ticket_id::text, author_id::text, body, is_internal, created_at
		FROM comments
		WHERE ticket_id = $1::uuid
		  AND ($2::boolean OR is_internal = FALSE)
		ORDER BY created_at ASC
		LIMIT $3 OFFSET $4
	`

	queryListCommentsNewest = `
		SELECT id::text, ticket_id::text, author_id::text, body, is_internal, created_at
		FROM comments
		WHERE ticket_id = $1::uuid
		  AND ($2::boolean OR is_internal = FALSE)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
)
