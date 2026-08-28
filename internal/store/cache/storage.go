package cache

import (
	"context"

	"github.com/loeia/gopherSocialNetwork/internal/store"
	"github.com/redis/go-redis/v9"
)

type User interface {
	Get(context.Context, int64) (*store.User, error)
	Set(context.Context, *store.User) error
	Delete(context.Context, int64) error
}

type Post interface {
	Get(context.Context, int64) (*store.Post, error)
	Set(context.Context, *store.Post) error
	Delete(context.Context, int64) error
}

type Avatar interface {
	Get(context.Context, int64) ([]byte, string, error)
	Set(context.Context, int64, []byte, string) error
	Delete(context.Context, int64) error
}

type Storage struct {
	User
	Post
	Avatar
}

func NewCacheStorage(rdb *redis.Client) *Storage {
	return &Storage{
		User:   &UserStore{rdb},
		Post:   &PostStore{rdb},
		Avatar: &AvatarStore{rdb},
	}
}
