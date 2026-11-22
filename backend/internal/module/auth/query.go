package auth

const (
	queryRegisterUser = `
		INSERT INTO users (name, email, phonenumber, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, name, email, phonenumber, created_at, updated_at, created_by::text
	`

	queryFindUserByEmail = `
		SELECT id::text, name, email, phonenumber, password_hash, created_at, updated_at, created_by::text
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`
)
