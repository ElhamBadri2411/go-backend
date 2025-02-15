package store

import (
	"context"      // Provides context management for request cancellation and timeouts
	"database/sql" // Standard library package for interacting with SQL databases
)

// `User` struct represents a user entity in the database.
//
// JSON struct tags (`json:"field_name"`) are included to ensure that this struct
// can be properly serialized/deserialized in API responses.
//
// Fields:
// - `ID` (int64): Unique identifier for the user.
// - `Username` (string): The user's username.
// - `Email` (string): The user's email address.
// - `Password` (string): The user's hashed password (excluded from JSON serialization).
// - `CreatedAt` (string): Timestamp indicating when the user was created.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"` // The `"-"` JSON tag ensures that `Password` is not included in JSON responses.
	CreatedAt string `json:"created_at"`
}

// `UsersRepositoryPostgres` is a concrete implementation of the `UsersRepository` interface.
// It interacts with a PostgreSQL database using an `sql.DB` connection pool.
type UsersRepositoryPostgres struct {
	db *sql.DB // Database connection pool
}

// `Create` inserts a new user into the `users` table and retrieves its assigned ID and creation timestamp.
//
// Parameters:
// - `ctx` (context.Context): Provides timeout and cancellation handling for the query.
// - `user` (*User): Pointer to a `User` struct containing user details.
//
// Returns:
// - An error if the query execution fails.
func (s *UsersRepositoryPostgres) Create(ctx context.Context, user *User) error {
	// SQL query to insert a new user into the database.
	// The `RETURNING` clause retrieves the newly created user's ID and creation timestamp.
	query := `
		INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id, created_at
	`

	// Execute the query and scan the returned values into the `user` struct.
	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Username, // Insert username
		user.Password, // Insert hashed password (should be hashed before insertion)
		user.Email,    // Insert email
	).Scan(
		&user.ID,        // Retrieve and store the newly generated user ID
		&user.CreatedAt, // Retrieve and store the creation timestamp
	)
	if err != nil {
		return err // Return the error if the insertion fails
	}

	return nil
}
