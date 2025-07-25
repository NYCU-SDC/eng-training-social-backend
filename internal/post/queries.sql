-- name: GetAll :many
SELECT * FROM posts;

-- name: GetByID :one
SELECT * FROM posts WHERE id = $1;

-- name: Create :one
INSERT INTO posts (title, content) VALUES ($1, $2) RETURNING *;

-- name: Update :one
UPDATE posts SET title = $1, content = $2, updated_at = now() WHERE id = $3 RETURNING *;

-- name: Delete :exec
DELETE FROM posts WHERE id = $1;