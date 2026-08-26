-- +goose Up
ALTER TABLE
    posts
ADD
    COLUMN view_count BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE
    posts
DROP COLUMN view_count;
