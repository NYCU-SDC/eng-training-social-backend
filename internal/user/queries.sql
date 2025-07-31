-- name: GetByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ExistsByEmail :one
SELECT EXISTS (
    SELECT 1 FROM users WHERE email = $1
) AS email_exists;

-- name: Create :one
INSERT INTO users (username, email) VALUES ($1, $2) RETURNING *;