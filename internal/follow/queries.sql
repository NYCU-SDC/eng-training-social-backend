-- name: ExistByID :one
SELECT EXISTS (
    SELECT 1 FROM following WHERE follower_id = $1 AND following_id = $2
) AS exists;

-- name: Create :one
INSERT INTO following (follower_id, following_id) VALUES ($1, $2)
ON CONFLICT (follower_id, following_id) DO NOTHING
RETURNING *;

-- name: Delete :exec
DELETE FROM following WHERE follower_id = $1 AND following_id = $2;