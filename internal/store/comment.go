package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Comment struct {
	ID              int64     `json:"id"`
	PostID          int64     `json:"post_id"`
	UserID          int64     `json:"user_id"`
	Content         string    `json:"content"`
	ParentID        *int64    `json:"parent_id"`
	ReplyToUserID   *int64    `json:"reply_to_user_id"`
	ReplyToUsername string    `json:"reply_to_username"`
	CreatedAt       time.Time `json:"created_at"`
	User            User      `json:"user"`
	LikeCount       int64     `json:"like_count"`
}

type CommentStore struct {
	db *sql.DB
}

func NewCommentStore(db *sql.DB) *CommentStore {
	return &CommentStore{
		db: db,
	}
}

func (s *CommentStore) GetByPostId(c context.Context, postId int64) ([]*Comment, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT 
			c.id,c.post_id,c.user_id,c.content,c.created_at,c.parent_id,c.reply_to_user_id,ru.username,u.username,c.like_count 
		FROM comments c
		JOIN users u on u.id = c.user_id
		LEFT JOIN users ru ON ru.id = c.reply_to_user_id
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC;
	`

	rows, err := s.db.QueryContext(ctx, query, postId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		var comment Comment
		var replyToUsername sql.NullString

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.ParentID,
			&comment.ReplyToUserID,
			&replyToUsername,
			&comment.User.Username,
			&comment.LikeCount,
		); err != nil {
			return nil, err
		}
		comment.ReplyToUsername = replyToUsername.String

		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *CommentStore) Create(ctx context.Context, comment *Comment) (*Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `INSERT INTO comments (user_id,post_id,content,parent_id,reply_to_user_id)
		VALUES ($1,$2,$3,$4,$5) RETURNING id,created_at`
	if err := s.db.QueryRowContext(
		ctx,
		query,
		comment.UserID,
		comment.PostID,
		comment.Content,
		comment.ParentID,
		comment.ReplyToUserID,
	).Scan(&comment.ID, &comment.CreatedAt); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentStore) Delete(c context.Context, commentId int64) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `DELETE FROM comments WHERE id = $1`

	if _, err := s.db.ExecContext(ctx, query, commentId); err != nil {
		return err
	}

	return nil
}

func (s *CommentStore) GetById(c context.Context, commentId int64) (*Comment, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT 
			c.id,c.post_id,c.user_id,c.content,c.created_at,c.parent_id,c.reply_to_user_id,ru.username,u.username,
			c.like_count 
		FROM comments c
		JOIN users u on u.id = c.user_id
		LEFT JOIN users ru ON ru.id = c.reply_to_user_id
		WHERE c.id = $1
	`

	var comment Comment
	var replyToUsername sql.NullString
	if err := s.db.QueryRowContext(ctx, query, commentId).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.ParentID,
		&comment.ReplyToUserID,
		&replyToUsername,
		&comment.User.Username,
		&comment.LikeCount,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	comment.ReplyToUsername = replyToUsername.String

	return &comment, nil
}
