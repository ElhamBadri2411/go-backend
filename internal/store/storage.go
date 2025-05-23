package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// `ErrNotFound` is a predefined error returned when a requested resource does not exist in the database.
var (
	ErrNotFound                 = errors.New("resource not found")
	ErrConflict                 = errors.New("duplicate key violates unique constraint")
	QueryContextTimeoutDuration = time.Second * 5
	ErrDuplicateEmail           = errors.New("there is already an account with that email")
	ErrDuplicateUsername        = errors.New("the username already exists")
)

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

	// `GetAll` retrieves posts from database (paginated) default offset 0, defulat limit 10
	GetAll(context.Context, int64, int64) ([]*Post, error)

	// `DeleteById` deletes a post given an id
	DeleteById(context.Context, int64) error

	// `UpdateById` updates a post given an id
	UpdateById(context.Context, *Post) error

	GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]*FeedPost, error)
}

type CommentsRepository interface {
	GetByPostId(context.Context, int64) ([]Comment, error)
	Create(context.Context, *Comment) error
}

// `UsersRepository` defines an interface for managing users in the database.
type UsersRepository interface {
	Create(context.Context, *sql.Tx, *User) error
	GetById(context.Context, int64) (*User, error)
	Follow(context.Context, int64, int64) error
	Unfollow(context.Context, int64, int64) error
	CreateAndInvite(context.Context, *User, string, time.Duration) error
	Activate(context.Context, string) error
	Delete(context.Context, int64) error
	GetByEmail(context.Context, string) (*User, error)
}

type RolesRepository interface {
	GetByName(context.Context, string) (*Role, error)
}

// `Storage` acts as a central repository abstraction layer.
// It embeds `PostsRepository` and `UsersRepository`, allowing unified access to database operations.
type Storage struct {
	PostsRepository    // Handles post-related database operations
	UsersRepository    // Handles user-related database operations
	CommentsRepository // Handles comment-related database operations
	RolesRepository    // Handles role-related database operations
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
		PostsRepository:    &PostsRepositoryPostgres{db}, // Instantiate PostgreSQL-backed posts repository
		UsersRepository:    &UsersRepositoryPostgres{db}, // Instantiate PostgreSQL-backed users repository
		CommentsRepository: &CommentRepositoryPostgres{db},
		RolesRepository:    &RoleRepositoryPostgres{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	opts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
