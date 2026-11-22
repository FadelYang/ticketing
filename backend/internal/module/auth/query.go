package auth

const (
	queryRegisterUser = `
		INSERT INTO users (name, email, phonenumber, role, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, name, email, phonenumber, role, created_at, updated_at, created_by::text
	`

	queryFindUserByEmail = `
		SELECT id::text, name, email, phonenumber, role, password_hash, created_at, updated_at, created_by::text
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`
)
