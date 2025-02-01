package store

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("resource not found")

type PostsRepository interface {
	Create(context.Context, *Post) error
	GetById(context.Context, int64) (*Post, error)
}

type UsersRepository interface {
	Create(context.Context, *User) error
}

type Storage struct {
	PostsRepository
	UsersRepository
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		PostsRepository: &PostsRepositoryPostgres{db},
		UsersRepository: &UsersRepositoryPostgres{db},
	}
}
