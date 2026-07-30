package store

import (
	"context"
	"database/sql"
)

type PostLikeStore struct {
	db *sql.DB
}

func NewPostLikeStore(db *sql.DB) *PostLikeStore {
	return &PostLikeStore{
		db: db,
	}
}

func (s *PostLikeStore) Like(c context.Context, postId, userId int64) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := "INSERT INTO post_likes (post_id,user_id) VALUES($1,$2) ON CONFLICT (post_id,user_id) DO NOTHING"

	if _, err := s.db.ExecContext(ctx, query, postId, userId); err != nil {
		return err
	}

	return nil
}

func (s *PostLikeStore) Dislike(c context.Context, postId, userId int64) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := "DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2"

	if _, err := s.db.ExecContext(ctx, query, postId, userId); err != nil {
		return err
	}

	return nil
}

func (s *PostLikeStore) GetPostLikes(c context.Context, postId int64) (int64, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := "SELECT COUNT(*) FROM post_likes WHERE post_id = $1"

	var count int64

	if err := s.db.QueryRowContext(ctx, query, postId).Scan(&count); err != nil {
		return -1, err
	}

	return count, nil
}
