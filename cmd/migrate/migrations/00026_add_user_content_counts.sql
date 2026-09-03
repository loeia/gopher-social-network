-- +goose Up
ALTER TABLE
    users
ADD
    COLUMN posts_count BIGINT DEFAULT 0,
ADD
    COLUMN likes_count BIGINT DEFAULT 0,
ADD
    COLUMN replies_count BIGINT DEFAULT 0;

UPDATE users u SET
    posts_count   = (SELECT COUNT(*) FROM posts p WHERE p.user_id = u.id),
    likes_count   = (SELECT COUNT(*) FROM post_likes l WHERE l.user_id = u.id),
    replies_count = (SELECT COUNT(*) FROM comments c WHERE c.user_id = u.id);

-- +goose StatementBegin
CREATE FUNCTION sync_user_posts_count() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE users SET posts_count = posts_count + 1 WHERE id = NEW.user_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE users SET posts_count = GREATEST(posts_count - 1, 0) WHERE id = OLD.user_id;
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_posts_sync_user_counts
    AFTER INSERT OR DELETE ON posts
    FOR EACH ROW EXECUTE FUNCTION sync_user_posts_count();

-- +goose StatementBegin
CREATE FUNCTION sync_user_likes_count() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE users SET likes_count = likes_count + 1 WHERE id = NEW.user_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE users SET likes_count = GREATEST(likes_count - 1, 0) WHERE id = OLD.user_id;
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_post_likes_sync_user_counts
    AFTER INSERT OR DELETE ON post_likes
    FOR EACH ROW EXECUTE FUNCTION sync_user_likes_count();

-- +goose StatementBegin
CREATE FUNCTION sync_user_replies_count() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE users SET replies_count = replies_count + 1 WHERE id = NEW.user_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE users SET replies_count = GREATEST(replies_count - 1, 0) WHERE id = OLD.user_id;
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_comments_sync_user_counts
    AFTER INSERT OR DELETE ON comments
    FOR EACH ROW EXECUTE FUNCTION sync_user_replies_count();

-- +goose Down
DROP TRIGGER IF EXISTS trg_posts_sync_user_counts ON posts;
DROP FUNCTION IF EXISTS sync_user_posts_count();
DROP TRIGGER IF EXISTS trg_post_likes_sync_user_counts ON post_likes;
DROP FUNCTION IF EXISTS sync_user_likes_count();
DROP TRIGGER IF EXISTS trg_comments_sync_user_counts ON comments;
DROP FUNCTION IF EXISTS sync_user_replies_count();
ALTER TABLE
    users
DROP COLUMN posts_count,
DROP COLUMN likes_count,
DROP COLUMN replies_count;
