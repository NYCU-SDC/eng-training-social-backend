CREATE TYPE reaction_type AS ENUM (
    'LIKE',
    'DISLIKE',
    'NONE'
);

CREATE TYPE content_type AS ENUM (
    'POST',
    'COMMENT'
);

CREATE TABLE IF NOT EXISTS reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction_type reaction_type NOT NULL,
    content_type content_type NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_reactions_post
    ON reactions (post_id, user_id)
    WHERE content_type = 'POST';

CREATE UNIQUE INDEX IF NOT EXISTS uq_reactions_comment
    ON reactions (comment_id, user_id)
    WHERE content_type = 'COMMENT';