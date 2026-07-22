-- +goose Up
CREATE TABLE IF NOT EXISTS user_invitations (
  token BYTEA PRIMARY KEY,
  user_id BIGINT NOT NULL,
  expiry TIMESTAMP(0) WITH TIME ZONE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS user_invitations;
