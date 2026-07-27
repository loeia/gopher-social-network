package cache

import "github.com/redis/go-redis/v9"

func NewRedisClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       db,
		Password: password,
	})
}
