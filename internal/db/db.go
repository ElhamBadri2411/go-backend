package db

import (
	"context"
	"database/sql" // Standard library package for interacting with SQL databases
	"time"
)

// `New` initi new database connn with given settings.
//
// - `url`: The PostgreSQL database connection string
// - `maxOpenConns`: The maximum number of open connections allowed in the connection pool.
// - `maxIdleConns`: The maximum number of idle connections allowed in the pool.
// - `maxIdleTime`: A duration string that specifies the maximum time an idle connection remains open.
//
// Returns a pointer to `sql.DB`, representing the database connection pool.

func New(url string, maxOpenConns int, maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	// Open a new database connection using the given PostgreSQL connection URL.
	// `sql.Open` does not establish a connection immediately; it validates the format of `url`.
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	// Parse `maxIdleTime` into a `time.Duration` object.
	// This determines how long idle connections remain open before being closed.
	duration, err := time.ParseDuration(maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Configure the database connection pool settings.
	db.SetMaxOpenConns(maxOpenConns) // Set the maximum number of open database connections
	db.SetMaxIdleConns(maxIdleConns) // Set the maximum number of idle connections
	db.SetConnMaxIdleTime(duration)  // Set the maximum idle time before a connection is closed

	// Create a timeout-based context for the initial database connection check.
	// This ensures the database is actually reachable before proceeding.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure the context is canceled to free up resources

	// `PingContext` checks if the database is reachable within the given timeout period.
	if err := db.PingContext(ctx); err != nil {
		return nil, err // Return an error if the database is unreachable
	}

	// Return the initialized database connection pool.
	return db, nil
}
