-- +goose Up
ALTER TABLE comments
    ADD COLUMN parent_id BIGINT REFERENCES comments(id) ON DELETE CASCADE,
    ADD COLUMN reply_to_user_id BIGINT REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX idx_comments_parent ON comments(parent_id);

-- +goose StatementBegin
CREATE FUNCTION sync_comment_count() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE posts SET comment_count = comment_count + 1 WHERE id = NEW.post_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = OLD.post_id;
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_comments_sync_count
    AFTER INSERT OR DELETE ON comments
    FOR EACH ROW EXECUTE FUNCTION sync_comment_count();

-- +goose StatementBegin
CREATE FUNCTION sync_like_count() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE posts SET like_count = like_count + 1 WHERE id = NEW.post_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE posts SET like_count = GREATEST(like_count - 1, 0) WHERE id = OLD.post_id;
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_post_likes_sync_count
    AFTER INSERT OR DELETE ON post_likes
    FOR EACH ROW EXECUTE FUNCTION sync_like_count();

-- +goose Down
DROP TRIGGER IF EXISTS trg_post_likes_sync_count ON post_likes;
DROP FUNCTION IF EXISTS sync_like_count();
DROP TRIGGER IF EXISTS trg_comments_sync_count ON comments;
DROP FUNCTION IF EXISTS sync_comment_count();
DROP INDEX IF EXISTS idx_comments_parent;
ALTER TABLE comments DROP COLUMN reply_to_user_id, DROP COLUMN parent_id;