package store

import (
	"context"
	"database/sql"
	"errors"
)

// `ErrNotFound` is a predefined error returned when a requested resource does not exist in the database.
var ErrNotFound = errors.New("resource not found")

// `PostsRepository` defines an interface for managing posts in the database.
// This interface enforces a contract that any implementation must adhere to.
type PostsRepository interface {
	// `Create` inserts a new post into the database.
	// - `context.Context`: Ensures request timeouts and cancellations are respected.
	// - `*Post`: Pointer to a `Post` struct containing post data.
	// Returns an error if the insertion fails.
	Create(context.Context, *Post) error

	// `GetById` retrieves a post from the database by its ID.
	// - `context.Context`: Ensures request timeouts and cancellations are respected.
	// - `int64`: The unique ID of the post.
	// Returns a pointer to the `Post` struct if found, otherwise an error.
	GetById(context.Context, int64) (*Post, error)
}

// `UsersRepository` defines an interface for managing users in the database.
type UsersRepository interface {
	// `Create` inserts a new user into the database.
	// - `context.Context`: Ensures request timeouts and cancellations are respected.
	// - `*User`: Pointer to a `User` struct containing user data.
	// Returns an error if the insertion fails.
	Create(context.Context, *User) error
}

// `Storage` acts as a central repository abstraction layer.
// It embeds `PostsRepository` and `UsersRepository`, allowing unified access to database operations.
type Storage struct {
	PostsRepository // Handles post-related database operations
	UsersRepository // Handles user-related database operations
}

// `NewStorage` initializes and returns a new `Storage` instance.
// It takes a `*sql.DB` object as input, which represents an active database connection.
//
// It creates two repository instances:
// - `PostsRepositoryPostgres`: Implements `PostsRepository` for PostgreSQL.
// - `UsersRepositoryPostgres`: Implements `UsersRepository` for PostgreSQL.
//
// These implementations interact with the database to perform CRUD operations.
func NewStorage(db *sql.DB) Storage {
	return Storage{
		PostsRepository: &PostsRepositoryPostgres{db}, // Instantiate PostgreSQL-backed posts repository
		UsersRepository: &UsersRepositoryPostgres{db}, // Instantiate PostgreSQL-backed users repository
	}
}
