-- +goose Up
DELETE FROM comments c WHERE NOT EXISTS (SELECT 1 FROM posts p WHERE p.id = c.post_id);
DELETE FROM comments c WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = c.user_id);

ALTER TABLE comments
    ADD CONSTRAINT fk_comments_post_id
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;

ALTER TABLE comments
    ADD CONSTRAINT fk_comments_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_post_id;
ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_user_id;
