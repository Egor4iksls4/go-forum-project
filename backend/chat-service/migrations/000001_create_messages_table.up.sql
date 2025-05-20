CREATE TABLE messages (
    id         SERIAL PRIMARY KEY,
    author     INTEGER                  NOT NULL,
    text       TEXT                     NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_author ON messages (author);
CREATE INDEX idx_messages_created_at ON messages (created_at);