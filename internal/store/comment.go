package store

import (
	"context"
	"database/sql"
	"errors"
)

type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
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
		SELECT c.id,c.post_id,c.user_id,c.content,c.created_at,u.username FROM comments c
		JOIN users u on u.id = c.user_id
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

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.User.Username,
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

func (s *CommentStore) Create(ctx context.Context, comment *Comment) (*Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `INSERT INTO comments (user_id,post_id,content) VALUES ($1,$2,$3) RETURNING id,created_at`
	if err := tx.QueryRowContext(ctx, query, comment.UserID, comment.PostID, comment.Content).Scan(&comment.ID, &comment.CreatedAt); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE posts SET comment_count = comment_count + 1 WHERE id = $1`, comment.PostID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommentStore) Delete(c context.Context, comment *Comment) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM comments WHERE id = $1`, comment.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = $1`, comment.PostID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *CommentStore) GetById(c context.Context, commentId int64) (*Comment, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT c.id,c.post_id,c.user_id,c.content,c.created_at,u.username FROM comments c
		JOIN users u on u.id = c.user_id
		WHERE c.id = $1
	`

	var comment Comment
	if err := s.db.QueryRowContext(ctx, query, commentId).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.User.Username,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &comment, nil
}
