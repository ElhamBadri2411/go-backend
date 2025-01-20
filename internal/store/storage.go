package store

import (
	"context"
	"database/sql"
)

type PostsRepository interface {
	Create(context.Context, *Post) error
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
		PostsRepository: &PostsStore{db},
		UsersRepository: &UsersStore{db},
	}
}
