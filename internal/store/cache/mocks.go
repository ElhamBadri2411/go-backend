package cache

import (
	"context"
	"log"

	"github.com/elhambadri2411/social/internal/store"
	"github.com/stretchr/testify/mock"
)

func NewMockCacheStore() Storage {
	return Storage{
		UsersCache: &MockUsersCacheRedis{},
	}
}

type MockUsersCacheRedis struct {
	mock.Mock
}

func (m *MockUsersCacheRedis) Get(ctx context.Context, id int64) (*store.User, error) {
	log.Println("calling mockUserCacheRedis.Get")
	args := m.Called(mock.Anything, id)
	return nil, args.Error(1)
}

func (m *MockUsersCacheRedis) Set(ctx context.Context, user *store.User) error {
	log.Println("calling mockUserCacheRedis.Set")
	args := m.Called(mock.Anything, user)
	return args.Error(0)
}
