package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID           int64    `json:"id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	UserID       int64    `json:"user_id"`
	Tags         []string `json:"tags"`
	CommentCount int64    `json:"comment_count"`
	LikeCount    int64    `json:"like_count"`
	Version      int64    `json:"version"`
	User         User     `json:"user"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostWithMetaData struct {
	Post         Post
	CommentCount int64
	LikeCount    int64
}

type SearchReq struct {
	Author string   `json:"author" validate:"max=100"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
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
	query := `
	SELECT
		p.id,p.user_id,p.title,p.content,p.tags,p.version,p.created_at,p.updated_at,u.username,
		p.comment_count,p.like_count
	FROM posts p
	JOIN users u ON p.user_id = u.id WHERE p.id = $1
	`

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
		&post.User.Username,
		&post.CommentCount,
		&post.LikeCount,
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
		    p.comment_count,
		    p.like_count
		FROM posts p
		LEFT JOIN users u ON p.user_id = u.id
		LEFT JOIN followers f ON f.user_id = p.user_id
		WHERE f.follower_id = $1
		AND (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
		AND (p.tags @> $5 OR $5 = '{}')
		AND (p.created_at >= $6::timestamptz OR $6 IS NULL)
		AND (p.created_at <= $7::timestamptz OR $7 IS NULL)
		ORDER BY p.created_at ` + pfq.Sort + `
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	var sinceArg any = nil
	if pfq.Since != "" {
		sinceArg = pfq.Since
	}

	var untilArg any = nil
	if pfq.Until != "" {
		untilArg = pfq.Until
	}

	rows, err := s.db.QueryContext(ctx, query, userId, pfq.Limit, pfq.Offset, pfq.Search, pq.Array(pfq.Tags), sinceArg, untilArg)
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
			&post.LikeCount,
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

func (s *PostStore) GetFree(c context.Context, count int) ([]*Post, error) {

	query := `
		SELECT p.id,p.user_id,p.title,p.content,p.tags,p.version,p.created_at,p.updated_at,p.comment_count,p.like_count,u.username
		FROM posts p JOIN users u ON p.user_id = u.id
		ORDER BY random()
		LIMIT $1
	`

	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			pq.Array(&post.Tags),
			&post.Version,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.CommentCount,
			&post.LikeCount,
			&post.User.Username,
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

// escapeLike escapes LIKE wildcards so user input matches literally.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// Search returns posts matching search/tags/author/date-range filters, ordered by relevance.
func (s *PostStore) Search(c context.Context, pfq *PaginatedFeedQuery) ([]*Post, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	var sinceArg any = nil
	if pfq.Since != "" {
		sinceArg = pfq.Since
	}
	var untilArg any = nil
	if pfq.Until != "" {
		untilArg = pfq.Until
	}

	order := "similarity(p.title, $8) DESC, p.created_at DESC"
	switch pfq.Sort {
	case "asc":
		order = "p.created_at ASC, similarity(p.title, $8) DESC"
	case "desc":
		order = "p.created_at DESC, similarity(p.title, $8) DESC"
	}

	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.tags, p.version, p.created_at,
		       p.comment_count, p.like_count, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE (p.title ILIKE '%' || $1 || '%' ESCAPE '\')
		  AND (p.tags @> $2)
		  AND (u.username ILIKE '%' || $3 || '%' ESCAPE '\')
		  AND (p.created_at >= $4::timestamptz OR $4 IS NULL)
		  AND (p.created_at <= $5::timestamptz OR $5 IS NULL)
		ORDER BY ` + order + `
		LIMIT $6 OFFSET $7
	`

	rows, err := s.db.QueryContext(ctx, query,
		escapeLike(pfq.Search),
		pq.Array(pfq.Tags),
		escapeLike(pfq.Author),
		sinceArg,
		untilArg,
		pfq.Limit,
		pfq.Offset,
		pfq.Search,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			pq.Array(&post.Tags),
			&post.Version,
			&post.CreatedAt,
			&post.CommentCount,
			&post.LikeCount,
			&post.User.Username,
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
