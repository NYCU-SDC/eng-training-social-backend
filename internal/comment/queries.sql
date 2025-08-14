-- name: GetAllByPostID :many
SELECT * FROM comments WHERE post_id = $1;

-- name: GetByID :one
SELECT * FROM comments WHERE id = $1;

-- name: Create :one
INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, $3) RETURNING *;

-- name: Update :one
UPDATE comments SET content = $1, updated_at = now() WHERE id = $2 RETURNING *;

-- name: Delete :exec
DELETE FROM comments WHERE id = $1;