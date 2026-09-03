-- +goose Up
ALTER TABLE followers
ADD CONSTRAINT followers_no_self_follow
CHECK (user_id <> follower_id);

-- +goose Down
ALTER TABLE followers
DROP CONSTRAINT followers_no_self_follow;
