CREATE TABLE IF NOT EXISTS comments (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    content text NOT NULL,
    author text NOT NULL,
    version integer NOT NULL DEFAULT 1
);
