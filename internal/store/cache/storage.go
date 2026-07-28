package cache

import (
	"context"

	"github.com/loeia/gopherSocialNetwork/internal/store"
	"github.com/redis/go-redis/v9"
)

type User interface {
	Get(context.Context, int64) (*store.User, error)
	Set(context.Context, *store.User) error
}

type Storage struct {
	User
}

func NewCacheStorage(rdb *redis.Client) *Storage {
	return &Storage{
		User: &UserStore{rdb},
	}
}
