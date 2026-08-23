package store

import (
	"context"
	"database/sql"
	"time"
)

type FavoriteCommentList struct {
	PostID     int64     `json:"post_id"`
	CommentID  int64     `json:"comment_id"`
	UserID     int64     `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	LikeCount  int64     `json:"like_count"`
	ReplyCount int64     `json:"reply_count"`
	Username   string    `json:"username"`
	Content    string    `json:"content"`
}

type CommentLikeStore struct {
	db *sql.DB
}

func NewCommentLikeStore(db *sql.DB) *CommentLikeStore {
	return &CommentLikeStore{
		db: db,
	}
}

func (s *CommentLikeStore) Like(c context.Context, commentId, userId int64) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
		INSERT INTO comment_likes (comment_id,user_id) VALUES ($1,$2)
		ON CONFLICT (comment_id,user_id) DO NOTHING
	`

	_, err := s.db.ExecContext(ctx, query, commentId, userId)
	return err
}

func (s *CommentLikeStore) Dislike(c context.Context, commentId, userId int64) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
		DELETE FROM comment_likes WHERE comment_id = $1 AND user_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, commentId, userId)
	return err
}

func (s *CommentLikeStore) GetUserFavoriteComments(c context.Context, userId int64) ([]*FavoriteCommentList, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
			SELECT
				c.post_id,c.id,c.like_count,u.id,u.username,l.created_at,c.content,
			(SELECT COUNT(*) FROM comments r WHERE r.parent_id = c.id) AS reply_count
			FROM comments c
			INNER JOIN comment_likes l ON c.id = l.comment_id
			LEFT JOIN users u ON c.user_id = u.id
			WHERE l.user_id = $1
			ORDER BY l.created_at DESC
		`

	rows, err := s.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*FavoriteCommentList, 0)
	for rows.Next() {
		var comment FavoriteCommentList

		if err := rows.Scan(
			&comment.PostID,
			&comment.CommentID,
			&comment.LikeCount,
			&comment.UserID,
			&comment.Username,
			&comment.CreatedAt,
			&comment.Content,
			&comment.ReplyCount,
		); err != nil {
			return nil, err
		}

		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
