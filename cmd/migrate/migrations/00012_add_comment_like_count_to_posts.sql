-- +goose Up
ALTER TABLE
    posts
ADD
    COLUMN comment_count BIGINT DEFAULT 0,
ADD
    COLUMN like_count BIGINT DEFAULT 0;

UPDATE posts p SET
    comment_count = (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id),
    like_count    = (SELECT COUNT(*) FROM post_likes l WHERE l.post_id = p.id);

-- +goose Down
ALTER TABLE
    posts
DROP COLUMN comment_count,
DROP COLUMN like_count;
