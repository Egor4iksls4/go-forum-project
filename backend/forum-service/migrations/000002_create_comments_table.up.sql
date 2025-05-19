CREATE TABLE comments (
    id         SERIAL PRIMARY KEY,
    post_id    INTEGER   NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    content    TEXT      NOT NULL,
    author     TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_post_id ON comments (post_id);