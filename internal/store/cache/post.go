package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/loeia/gopherSocialNetwork/internal/store"
	"github.com/redis/go-redis/v9"
)

const PostExpTime = time.Minute * 5

type PostStore struct {
	rdb *redis.Client
}

func (s *PostStore) Get(c context.Context, postId int64) (*store.Post, error) {
	cacheKey := fmt.Sprintf("post-%v", postId)

	data, err := s.rdb.Get(c, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var post store.Post
	if data != "" {
		if err := json.Unmarshal([]byte(data), &post); err != nil {
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostStore) Set(c context.Context, post *store.Post) error {
	cacheKey := fmt.Sprintf("post-%v", post.ID)

	jsonData, err := json.Marshal(post)
	if err != nil {
		return err
	}

	return s.rdb.SetEx(c, cacheKey, jsonData, PostExpTime).Err()
}

func (s *PostStore) Delete(c context.Context, postId int64) error {
	cacheKey := fmt.Sprintf("post-%v", postId)
	return s.rdb.Del(c, cacheKey).Err()
}
