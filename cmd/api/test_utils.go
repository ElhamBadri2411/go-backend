package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elhambadri2411/social/internal/auth"
	"github.com/elhambadri2411/social/internal/store"
	"github.com/elhambadri2411/social/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()
	logger := zap.Must(zap.NewProduction()).Sugar()
	// logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockCacheStore()
	mockAuthenticator := auth.NewMockJWTAuthenticator()

	return &application{
		store:         mockStore,
		logger:        logger,
		cache:         mockCacheStore,
		authenticator: mockAuthenticator,
	}
}

func execRequest(req *http.Request, mux *chi.Mux) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func assertResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response %d. Got %d", expected, actual)
	}
}
