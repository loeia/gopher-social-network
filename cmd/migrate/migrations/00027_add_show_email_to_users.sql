-- +goose Up
ALTER TABLE users ADD COLUMN show_email BOOLEAN NOT NULL DEFAULT false;
-- +goose Down
ALTER TABLE users DROP COLUMN show_email;
