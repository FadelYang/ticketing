package attachment

const (
	queryCreateAttachment = `
		INSERT INTO attachments (type, target_id, uploader_id, file_name, content_type, size_bytes, object_key)
		VALUES ($1, $2::uuid, $3::uuid, $4, $5, $6, $7)
		RETURNING id::text, type, target_id::text, uploader_id::text, file_name, content_type, size_bytes, object_key, created_at
	`

	queryListAttachmentsByTargets = `
		SELECT id::text, type, target_id::text, uploader_id::text, file_name, content_type, size_bytes, object_key, created_at
		FROM attachments
		WHERE type = $1
		  AND target_id = ANY($2::uuid[])
		ORDER BY created_at ASC
	`

	queryGetAttachmentByID = `
		SELECT id::text, type, target_id::text, uploader_id::text, file_name, content_type, size_bytes, object_key, created_at
		FROM attachments
		WHERE id = $1::uuid
	`
)
