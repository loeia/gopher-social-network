package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const AvatarExpTime = time.Minute * 10

type AvatarStore struct {
	rdb *redis.Client
}

type avatarData struct {
	Data []byte `json:"data"`
	Mime string `json:"mime"`
}

func (s *AvatarStore) Get(c context.Context, userId int64) ([]byte, string, error) {
	cacheKey := fmt.Sprintf("avatar-%v", userId)

	data, err := s.rdb.Get(c, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, "", nil
		}
		return nil, "", err
	}

	var av avatarData
	if data != "" {
		if err := json.Unmarshal([]byte(data), &av); err != nil {
			return nil, "", err
		}
	}
	return av.Data, av.Mime, nil
}

func (s *AvatarStore) Set(c context.Context, userId int64, data []byte, mime string) error {
	cacheKey := fmt.Sprintf("avatar-%v", userId)

	av := avatarData{Data: data, Mime: mime}
	jsonData, err := json.Marshal(av)
	if err != nil {
		return err
	}

	return s.rdb.SetEx(c, cacheKey, jsonData, AvatarExpTime).Err()
}

func (s *AvatarStore) Delete(c context.Context, userId int64) error {
	cacheKey := fmt.Sprintf("avatar-%v", userId)
	return s.rdb.Del(c, cacheKey).Err()
}
