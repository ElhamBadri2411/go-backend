package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elhambadri2411/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type UsersCacheRedis struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute * 5

func (s *UsersCacheRedis) Get(ctx context.Context, userId int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userId)
	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (s *UsersCacheRedis) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.rdb.SetEX(ctx, cacheKey, userJson, UserExpTime).Err()
}
