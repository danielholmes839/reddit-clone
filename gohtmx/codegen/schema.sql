CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(300) NOT NULL UNIQUE,
    password BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    root_id UUID NOT NULL,
    parent_id UUID,
    author_id INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    title TEXT,
    content TEXT NOT NULL,
    FOREIGN KEY (author_id) REFERENCES users(id),
    FOREIGN KEY (root_id) REFERENCES threads(id),
    FOREIGN KEY (parent_id) REFERENCES threads(id)
);

CREATE TABLE IF NOT EXISTS votes (
    id SERIAL PRIMARY KEY,
    voter_id INT NOT NULL,
    thread_id UUID NOT NULL,
    upvote BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    FOREIGN KEY (voter_id) REFERENCES users(id),
    FOREIGN KEY (thread_id) REFERENCES threads(id),
    UNIQUE(voter_id, thread_id)
);

-- user indexes
CREATE INDEX ON users (username);

-- vote indexes
CREATE INDEX ON votes (thread_id);

CREATE INDEX ON votes (upvote);

-- thread indexes
CREATE INDEX ON threads (author_id);

CREATE INDEX ON threads (root_id);

CREATE INDEX ON threads (parent_id);

CREATE INDEX ON threads (created_at);