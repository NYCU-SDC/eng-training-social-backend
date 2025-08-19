-- name: CreateByPostID :one
INSERT INTO reactions (post_id, user_id, reaction_type, content_type)
VALUES ($1, $2, $3, 'POST')
ON CONFLICT (post_id, user_id)
    WHERE content_type = 'POST'
    DO UPDATE SET reaction_type = EXCLUDED.reaction_type, updated_at = now()
RETURNING *;

-- name: ExistsByPostIDAndUserID :one
SELECT EXISTS (
    SELECT 1 FROM reactions WHERE post_id = $1 AND user_id = $2 AND content_type = 'POST'
) AS exists;

-- name: GetByPostIDAndUserID :one
SELECT reactions.reaction_type FROM reactions WHERE post_id = $1 AND user_id = $2;

-- name: DeleteByPostIDAndUserID :exec
DELETE FROM reactions WHERE post_id = $1 AND user_id = $2;

-- name: CreateByCommentID :one
INSERT INTO reactions (comment_id, user_id, reaction_type, content_type)
VALUES ($1, $2, $3, 'COMMENT')
ON CONFLICT (comment_id, user_id)
    WHERE content_type = 'COMMENT'
    DO UPDATE SET reaction_type = EXCLUDED.reaction_type, updated_at = now()
RETURNING *;

-- name: ExistsByCommentIDAndUserID :one
SELECT EXISTS (
    SELECT 1 FROM reactions WHERE comment_id = $1 AND user_id = $2 AND content_type = 'COMMENT'
) AS exists;

-- name: GetByCommentIDAndUserID :one
SELECT reactions.reaction_type FROM reactions WHERE comment_id = $1 AND user_id = $2;

-- name: DeleteByCommentIDAndUserID :exec
DELETE FROM reactions WHERE comment_id = $1 AND user_id = $2;