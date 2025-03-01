package store

import (
	"context"      // Provides context management for request cancellation and timeouts
	"database/sql" // Standard library package for interacting with SQL databases
	"errors"       // Standard library package for defining and handling errors
	"log"          // Standard library package for logging errors and informational messages

	// PostgreSQL driver for Go, which includes utilities for handling PostgreSQL-specific data types.
	// `pq.Array` is used to handle array data types in PostgreSQL.
	"github.com/lib/pq"
)

// `Post` struct represents a post entity in the database.
//
// JSON struct tags (`json:"field_name"`) are included to ensure that this struct
// can be serialized/deserialized properly when used in APIs.
//
// Fields:
// - `ID` (int64): Unique identifier of the post.
// - `Content` (string): The body/content of the post.
// - `Title` (string): Title of the post.
// - `UserId` (int64): ID of the user who created the post.
// - `Tags` ([]string): A list of tags associated with the post (stored as an array in PostgreSQL).
// - `CreatedAt` (string): Timestamp when the post was created.
// - `UpdatedAt` (string): Timestamp when the post was last updated.
type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserId    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comments"`
}

// `PostsRepositoryPostgres` is a concrete implementation of the `PostsRepository` interface.
// It interacts with a PostgreSQL database using an `sql.DB` connection pool.
type PostsRepositoryPostgres struct {
	db *sql.DB // Database connection pool
}

// `Create` inserts a new post into the `posts` table and retrieves its assigned ID and timestamps.
//
// Parameters:
// - `ctx` (context.Context): Provides timeout and cancellation handling for the query.
// - `post` (*Post): Pointer to a `Post` struct containing post details.
//
// Returns:
// - An error if the query execution fails.
func (s *PostsRepositoryPostgres) Create(ctx context.Context, post *Post) error {
	// SQL query to insert a new post into the database.
	// The `RETURNING` clause retrieves the newly created post's ID, creation timestamp, and update timestamp.
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`

	// Execute the query and scan the returned values into the `post` struct.
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserId,
		pq.Array(post.Tags), // Converts Go slice to a PostgreSQL array
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		log.Println(err.Error()) // Log the error for debugging
		return err
	}

	return nil
}

// `GetById` retrieves a post from the `posts` table by its unique ID.
//
// Parameters:
// - `ctx` (context.Context): Provides timeout and cancellation handling for the query.
// - `id` (int64): The unique identifier of the post.
//
// Returns:
// - A pointer to a `Post` struct if found.
// - `ErrNotFound` if no matching post is found.
// - An error if the query execution fails.
func (s *PostsRepositoryPostgres) GetById(ctx context.Context, id int64) (*Post, error) {
	var post Post // Struct to hold the retrieved post data

	// SQL query to fetch a post by its ID.
	query := `
		SELECT id, user_id, title, content, created_at, updated_at, tags FROM posts WHERE id = $1
	`

	// Execute the query and scan the result into the `post` struct.
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags), // Converts PostgreSQL array to Go slice
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// If no row is found, return a predefined `ErrNotFound` error.
			return nil, ErrNotFound
		default:
			// Return any other database-related errors.
			return nil, err
		}
	}

	return &post, nil
}
