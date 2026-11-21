package example

// Raw SQL queries for the example module.
// Kept separate from repository.go for clarity and maintainability.

const (
	queryCreateExample = `
		INSERT INTO examples (name, description)
		VALUES ($1, $2)
		RETURNING id::text, name, description, created_at, updated_at
	`

	queryGetExampleByID = `
		SELECT id::text, name, description, created_at, updated_at
		FROM examples
		WHERE id = $1::uuid
	`

	queryListExamples = `
		SELECT id::text, name, description, created_at, updated_at
		FROM examples
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
	`

	queryUpdateExample = `
		UPDATE examples
		SET name = $2,
		    description = $3,
		    updated_at = NOW()
		WHERE id = $1::uuid
		RETURNING id::text, name, description, created_at, updated_at
	`

	queryDeleteExample = `
		DELETE FROM examples
		WHERE id = $1::uuid
	`
)
