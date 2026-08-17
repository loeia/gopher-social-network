package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type FavoritePostList struct {
	PostID       int64     `json:"post_id"`
	Author       string    `json:"author"`
	Title        string    `json:"title"`
	Tags         []string  `json:"tags"`
	CommentCount int64     `json:"comment_count"`
	LikeCount    int64     `json:"like_count"`
	CreatedAt    time.Time `json:"created_at"`
}

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

func (s *PostLikeStore) GetUserFavoritePosts(c context.Context, userId int64) ([]*FavoritePostList, error) {

	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT p.id,u.username,p.title,p.tags,p.comment_count,p.like_count,p.created_at FROM posts p
		LEFT JOIN post_likes l ON p.id = l.post_id
		LEFT JOIN users u ON p.user_id = u.id
		WHERE l.user_id = $1
		ORDER BY l.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*FavoritePostList, 0)
	for rows.Next() {
		var post FavoritePostList

		if err := rows.Scan(
			&post.PostID,
			&post.Author,
			&post.Title,
			pq.Array(&post.Tags),
			&post.CommentCount,
			&post.LikeCount,
			&post.CreatedAt,
		); err != nil {
			return nil, err
		}

		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}
