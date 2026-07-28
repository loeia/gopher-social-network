-- +goose Up
ALTER TABLE
  users
ADD
  COLUMN is_active BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE users DROP is_active;
