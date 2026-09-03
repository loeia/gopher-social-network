-- +goose Up
ALTER TABLE
    users
ADD
    COLUMN followers_count BIGINT DEFAULT 0,
ADD
    COLUMN following_count BIGINT DEFAULT 0;

UPDATE users u SET
    followers_count = (SELECT COUNT(*) FROM followers f WHERE f.user_id = u.id),
    following_count = (SELECT COUNT(*) FROM followers f WHERE f.follower_id = u.id);

-- +goose StatementBegin
CREATE FUNCTION sync_follow_counts() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE users SET followers_count = followers_count + 1 WHERE id = NEW.user_id;
        UPDATE users SET following_count = following_count + 1 WHERE id = NEW.follower_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE users SET followers_count = GREATEST(followers_count - 1, 0) WHERE id = OLD.user_id;
        UPDATE users SET following_count = GREATEST(following_count - 1, 0) WHERE id = OLD.follower_id;
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_followers_sync_counts
    AFTER INSERT OR DELETE ON followers
    FOR EACH ROW EXECUTE FUNCTION sync_follow_counts();

-- +goose Down
DROP TRIGGER IF EXISTS trg_followers_sync_counts ON followers;
DROP FUNCTION IF EXISTS sync_follow_counts();
ALTER TABLE
    users
DROP COLUMN followers_count,
DROP COLUMN following_count;
