-- +goose Up
CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       email TEXT UNIQUE NOT NULL,
                       password_hash TEXT NOT NULL,
                       created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;