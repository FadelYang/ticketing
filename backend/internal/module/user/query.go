package user

const (
	queryCreateUser = `
		INSERT INTO users (name, email, phonenumber, role, created_by, password_hash)
		VALUES ($1, $2, $3, $4, $5::uuid, $6)
		RETURNING id::text, name, email, phonenumber, role, created_at, updated_at, created_by::text
	`

	queryGetUserByID = `
		SELECT id::text, name, email, phonenumber, role, created_at, updated_at, created_by::text
		FROM users
		WHERE id = $1::uuid
	`

	queryListUsers = `
		SELECT id::text, name, email, phonenumber, role, created_at, updated_at, created_by::text
		FROM users
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%' OR phonenumber ILIKE '%' || $1 || '%' OR role ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
	`

	queryUpdateUser = `
		UPDATE users
		SET name = $2,
		    email = $3,
		    phonenumber = $4,
		    role = COALESCE(NULLIF($5, ''), role),
		    updated_at = NOW()
		WHERE id = $1::uuid
		RETURNING id::text, name, email, phonenumber, role, created_at, updated_at, created_by::text
	`

	queryUpdateUserRole = `
		UPDATE users
		SET role = $2,
		    updated_at = NOW()
		WHERE id = $1::uuid
		RETURNING id::text, name, email, phonenumber, role, created_at, updated_at, created_by::text
	`

	queryDeleteUser = `
		DELETE FROM users
		WHERE id = $1::uuid
	`
)
