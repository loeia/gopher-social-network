-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users(
  id BIGSERIAL PRIMARY KEY,
  email citext UNIQUE NOT NULL,
  username VARCHAR(255) UNIQUE NOT NULL,
  password bytea NOT NULL,
  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS citext;
