-- +goose Up
CREATE TABLE links (
    alias TEXT PRIMARY KEY,
    url TEXT NOT NULL,
    expires_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE links;
