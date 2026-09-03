package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

func NewRedisRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		rdb:    rdb,
		limit:  limit,
		window: window,
	}
}

func (rl *RedisRateLimiter) Allow(ip string) (bool, time.Duration) {
	ctx := context.Background()
	key := fmt.Sprintf("ratelimit:%s", ip)

	count, err := rl.rdb.Incr(ctx, key).Result()
	if err != nil {
		return true, 0
	}

	if count == 1 {
		rl.rdb.Expire(ctx, key, rl.window)
	}

	if count > int64(rl.limit) {
		ttl, err := rl.rdb.TTL(ctx, key).Result()
		if err != nil || ttl <= 0 {
			return false, rl.window
		}
		return false, ttl
	}

	return true, 0
}
