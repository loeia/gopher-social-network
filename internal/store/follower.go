package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type Follower struct {
	UserID     int64     `json:"user_id"`
	FollowerID int64     `json:"follower_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func NewFollowerStore(db *sql.DB) *FollowerStore {
	return &FollowerStore{
		db: db,
	}
}

// userId: 		People who are being followed
// followerId:  People who pay attention to others
func (s FollowerStore) Follow(c context.Context, userId, followerId int64) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := "INSERT INTO followers (user_id,follower_id) VALUES ($1,$2)"

	res, err := s.db.ExecContext(ctx, query, userId, followerId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrFollowFailed
	}

	return nil
}

// userId: 		People who want to be unfollowed
// followerId:  Cancel following someone else's followers
func (s *FollowerStore) Unfollow(c context.Context, userId, followerId int64) error {
	query := `DELETE FROM followers WHERE user_id = $1 AND follower_id = $2`

	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	if _, err := s.db.ExecContext(c, query, userId, followerId); err != nil {
		return err
	}

	return nil
}
