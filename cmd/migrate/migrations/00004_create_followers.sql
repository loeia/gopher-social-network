-- +goose Up
CREATE TABLE IF NOT EXISTS followers (
  user_id BIGINT NOT NULL,
  follower_id BIGINT NOT NULL,
  created_at TIMESTAMP(0) WITH TIME ZONE DEFAULT NOW(),

  PRIMARY KEY (user_id,follower_id),
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  FOREIGN KEY (follower_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS followers;
