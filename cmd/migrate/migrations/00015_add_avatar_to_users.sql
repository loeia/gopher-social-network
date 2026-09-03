-- +goose Up
ALTER TABLE users ADD COLUMN avatar BYTEA;
ALTER TABLE users ADD COLUMN avatar_mime VARCHAR(50);
-- +goose Down
ALTER TABLE users DROP COLUMN avatar;
ALTER TABLE users DROP COLUMN avatar_mime;
