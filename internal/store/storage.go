package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const (
	QueryTimeoutDuration = time.Second * 5
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

type PostStorage interface {
	Create(context.Context, *Post) error
	GetById(context.Context, int64) (*Post, error)
	Delete(context.Context, *Post) error
	Update(context.Context, *Post) error
}
type UserStorage interface {
	Create(context.Context, *User) error
	GetById(context.Context, int64) (*User, error)
}
type CommentStorage interface {
	Create(context.Context, *Comment) error
	GetById(context.Context, int64) ([]*Comment, error)
}

type Storage struct {
	Posts    PostStorage
	Users    UserStorage
	Comments CommentStorage
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Posts:    NewPostStore(db),
		Users:    NewUserStore(db),
		Comments: NewCommentStore(db),
	}
}
