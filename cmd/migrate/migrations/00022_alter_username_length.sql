-- +goose Up
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(25);

-- +goose Down
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(255);
