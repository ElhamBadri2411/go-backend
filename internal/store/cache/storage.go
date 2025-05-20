package cache

import (
	"context"

	"github.com/elhambadri2411/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type UsersCache interface {
	Get(context.Context, int64) (*store.User, error)
	Set(context.Context, *store.User) error
}

type Storage struct {
	UsersCache
}

func NewCacheStorage(rdb *redis.Client) Storage {
	return Storage{
		UsersCache: &UsersCacheRedis{rdb: rdb},
	}
}
