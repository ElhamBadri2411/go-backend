package main

import (
	"log"
	"net/http"
	"testing"

	"github.com/elhambadri2411/social/internal/store/cache"
	"github.com/stretchr/testify/mock"
)

func TestGetUser(t *testing.T) {
	mockApp := newTestApplication(t)
	mockMux := mockApp.mount()

	testToken, _ := mockApp.authenticator.GenerateToken(nil)

	t.Run("should not allow unauthenticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := execRequest(req, mockMux)
		assertResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow authenticated requests", func(t *testing.T) {
		mockCacheStore := mockApp.cache.UsersCache.(*cache.MockUsersCacheRedis)

		mockCacheStore.On("Get", mock.Anything, int64(1)).Return(nil, nil)
		mockCacheStore.On("Get", mock.Anything, int64(21)).Return(nil, nil)
		mockCacheStore.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := execRequest(req, mockMux)
		assertResponseCode(t, http.StatusOK, rr.Code)
		mockCacheStore.Calls = nil
	})

	t.Run("should hit cache first and if not exists sets user on the cache", func(t *testing.T) {
		mockCacheStore := mockApp.cache.UsersCache.(*cache.MockUsersCacheRedis)

		mockCacheStore.On("Get", mock.Anything, int64(1)).Return(nil, nil)  // user we are getting
		mockCacheStore.On("Get", mock.Anything, int64(21)).Return(nil, nil) // user that is logged in (form jwt)
		mockCacheStore.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)
		rr := execRequest(req, mockMux)

		assertResponseCode(t, http.StatusOK, rr.Code)

		for _, call := range mockCacheStore.Calls {
			log.Println("Called method", call.Method)
		}
		mockCacheStore.AssertNumberOfCalls(t, "Get", 2)
		mockCacheStore.Calls = nil
	})
}
