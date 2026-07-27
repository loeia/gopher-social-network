package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID       int64      `json:"id"`
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	UserID   int64      `json:"user_id"`
	Tags     []string   `json:"tags"`
	Comments []*Comment `json:"comments"`
	Version  int64      `json:"version"`
	User     User       `json:"user"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type PostWithMetaData struct {
	Post         Post  `json:"post"`
	CommentCount int64 `json:"comment_count"`
}

type PostStore struct {
	db *sql.DB
}

func NewPostStore(db *sql.DB) *PostStore {
	return &PostStore{
		db: db,
	}
}

func (s *PostStore) Create(c context.Context, post *Post) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
		INSERT INTO posts (title,content,user_id,tags)
		VALUES ($1,$2,$3,$4) RETURNING id,created_at,updated_at
	`

	row := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.UserID, pq.Array(post.Tags))
	err := row.Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetById(c context.Context, id int64) (*Post, error) {
	query := `SELECT id,user_id,title,content,tags,version,created_at,updated_at FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	row := s.db.QueryRowContext(ctx, query, id)

	var post Post
	err := row.Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostStore) Delete(ctx context.Context, post *Post) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, post.ID)
	if err != nil {
		return err
	}

	row, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if row == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostStore) Update(c context.Context, post *Post) error {
	query := `UPDATE posts SET title = $1, content = $2, tags = $3,updated_at = NOW(),version = version + 1 WHERE id = $4 AND version = $5 RETURNING version,updated_at`

	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
		post.ID,
		post.Version,
	).Scan(&post.Version, &post.UpdatedAt); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *PostStore) GetUserFeed(c context.Context, userId int64, pfq *PaginatedFeedQuery) ([]*PostWithMetaData, error) {
	query := `
		SELECT
		    p.id,
		    p.user_id,
		    p.title,
		    p.content,
		    p.created_at,
		    p.version,
		    p.tags,
		    u.username,
		    COUNT(c.id) AS comments_count
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		LEFT JOIN followers f ON f.user_id = p.user_id
		WHERE f.follower_id = $1
		AND (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
		AND (p.tags @> $5 OR $5 = '{}')
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + pfq.Sort + `
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userId, pfq.Limit, pfq.Offset, pfq.Search, pq.Array(pfq.Tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []*PostWithMetaData
	for rows.Next() {
		var post PostWithMetaData
		if err := rows.Scan(
			&post.Post.ID,
			&post.Post.UserID,
			&post.Post.Title,
			&post.Post.Content,
			&post.Post.CreatedAt,
			&post.Post.Version,
			pq.Array(&post.Post.Tags),
			&post.Post.User.Username,
			&post.CommentCount,
		); err != nil {
			return nil, err
		}

		feed = append(feed, &post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feed, nil
}
