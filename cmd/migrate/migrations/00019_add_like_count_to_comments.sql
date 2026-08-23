-- +goose Up
ALTER TABLE comments ADD COLUMN like_count BIGINT DEFAULT 0;

UPDATE comments c SET like_count = (
    SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id
);

-- +goose StatementBegin
CREATE FUNCTION sync_comment_like_count() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE comments SET like_count = like_count + 1 WHERE id = NEW.comment_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE comments SET like_count = GREATEST(like_count - 1, 0) WHERE id = OLD.comment_id;
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_comment_likes_sync_count
    AFTER INSERT OR DELETE ON comment_likes
    FOR EACH ROW EXECUTE FUNCTION sync_comment_like_count();

-- +goose Down
DROP TRIGGER IF EXISTS trg_comment_likes_sync_count ON comment_likes;
DROP FUNCTION IF EXISTS sync_comment_like_count();
ALTER TABLE comments DROP COLUMN like_count;
