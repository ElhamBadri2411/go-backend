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
	Version   int64     `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type FeedPost struct {
	Post
	CommentCount int `json:"comments_count"`
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

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

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
		SELECT id, user_id, title, content, created_at, updated_at, tags, version FROM posts WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()
	// Execute the query and scan the result into the `post` struct.
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags), // Converts PostgreSQL array to Go slice
		&post.Version,
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

// `GetAll` retrieves a post from the `posts` table by its unique ID.
//
// Parameters:
// - `ctx` (context.Context): Provides timeout and cancellation handling for the query.
// - `limit` (int64): for pagination
// - `offset` (int64): for pagination
//
// Returns:
// - A pointer to posts array
// - `ErrNotFound` if no matching post is found.
// - An error if the query execution fails.
func (s *PostsRepositoryPostgres) GetAll(ctx context.Context, limit int64, offset int64) ([]*Post, error) {
	var posts []*Post

	query := `
		SELECT id, user_id, title, content, created_at, updated_at, tags, version FROM posts ORDER BY id LIMIT $1 OFFSET $2;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post

		err := rows.Scan(
			&post.ID,
			&post.UserId,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			pq.Array(&post.Tags), // Converts PostgreSQL array to Go slice
			&post.Version,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (s *PostsRepositoryPostgres) DeleteById(ctx context.Context, id int64) error {
	query := `
		DELETE FROM posts WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()
	// Execute the query and scan the result into the `post` struct.
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// If no row is found, return a predefined `ErrNotFound` error.
			return ErrNotFound
		default:
			// Return any other database-related errors.
			return err
		}
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return ErrNotFound
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostsRepositoryPostgres) UpdateById(ctx context.Context, post *Post) error {
	// SQL query to fetch a post by its ID.
	query := `
		UPDATE posts
		SET title = $1, content = $2, version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.ID,
		post.Version,
	).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		}
	}

	return nil
}

func (s *PostsRepositoryPostgres) GetUserFeed(ctx context.Context, userId int64, pfq PaginatedFeedQuery) ([]*FeedPost, error) {
	query := `
		SELECT p.id, p.title, p."content", p.user_id, p.created_at, p.tags, u.username, COUNT(c.id) AS comments_count
		FROM posts p 
		LEFT JOIN "comments" c ON c.post_id=p.id
		LEFT JOIN users u  ON  p.user_id = u.id 
		JOIN followers f ON f.user_id = p.user_id or p.user_id = $1
		WHERE f.follower_id = $1
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + pfq.Sort + ` 
		LIMIT $2 OFFSET $3;
	`
	var feedPosts []*FeedPost

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userId, pfq.Limit, pfq.Offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var feedPost FeedPost

		err := rows.Scan(
			&feedPost.ID,
			&feedPost.Title,
			&feedPost.Content,
			&feedPost.UserId,
			&feedPost.CreatedAt,
			pq.Array(&feedPost.Tags), // Converts PostgreSQL array to Go slice
			&feedPost.User.Username,
			&feedPost.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		feedPosts = append(feedPosts, &feedPost)
	}
	return feedPosts, nil
}
