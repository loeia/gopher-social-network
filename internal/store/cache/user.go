package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/loeia/gopherSocialNetwork/internal/store"
	"github.com/redis/go-redis/v9"
)

const UserExpTime = time.Minute

type UserStore struct {
	rdb *redis.Client
}

func (s *UserStore) Get(c context.Context, userId int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userId)

	data, err := s.rdb.Get(c, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var user store.User
	if data != "" {
		if err := json.Unmarshal([]byte(data), &user); err != nil {
			return nil, err
		}
	}
	return &user, nil
}

func (s *UserStore) Set(c context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	jsonData, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.rdb.SetEx(c, cacheKey, jsonData, UserExpTime).Err()
}

func (s *UserStore) Delete(c context.Context, userId int64) error {
	cacheKey := fmt.Sprintf("user-%v", userId)
	return s.rdb.Del(c, cacheKey).Err()
}
