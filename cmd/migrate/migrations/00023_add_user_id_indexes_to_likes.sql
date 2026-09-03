-- +goose Up
CREATE INDEX idx_post_likes_user_id ON post_likes(user_id);
CREATE INDEX idx_comment_likes_user_id ON comment_likes(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_post_likes_user_id;
DROP INDEX IF EXISTS idx_comment_likes_user_id;
