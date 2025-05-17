package store

import (
	"context" // Provides context management for request cancellation and timeouts
	"crypto/sha256"
	"database/sql" // Standard library package for interacting with SQL databases
	"encoding/hex"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"password"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleId    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerId int64  `json:"follower_id"`
	Created_at string `json:"created_at"`
}

type Password struct {
	text *string
	hash []byte
}

func (p *Password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
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
func (s *UsersRepositoryPostgres) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	// SQL query to insert a new user into the database.
	// The `RETURNING` clause retrieves the newly created user's ID and creation timestamp.
	query := `
		INSERT INTO users (username, password, email, role_id) VALUES ($1, $2, $3, (SELECT ID FROM roles WHERE name = $4)) RETURNING id, created_at
	`

	role := user.Role.Name
	if role == "" {
		role = "user"
	}

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()
	// Execute the query and scan the returned values into the `user` struct.
	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,      // Insert username
		user.Password.hash, // Insert hashed password (should be hashed before insertion)
		user.Email,         // Insert email
		role,
	).Scan(
		&user.ID,        // Retrieve and store the newly generated user ID
		&user.CreatedAt, // Retrieve and store the creation timestamp
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *UsersRepositoryPostgres) CreateAndInvite(ctx context.Context, user *User, token string, invitation_expiry time.Duration) error {
	// transaction wrapper
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// create user
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}
		// create invite
		if err := s.createUserInvitation(ctx, tx, token, user.ID, invitation_expiry); err != nil {
			return err
		}

		return nil
	})
}

func (s *UsersRepositoryPostgres) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, userId int64, invitation_expiry time.Duration) error {
	query := `
		INSERT INTO user_invitations (user_id, token, expiry) VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId, token, time.Now().Add(invitation_expiry))
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersRepositoryPostgres) GetById(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT users.id, email, username, password, created_at, roles.*
		FROM users
		JOIN roles
		ON users.role_id = roles.id
		WHERE users.id = $1 AND is_active = true
	`
	var user User
	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UsersRepositoryPostgres) Follow(ctx context.Context, userToFollowId int64, userId int64) error {
	query := `
		INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()
	// Execute the query and scan the returned values into the `user` struct.
	_, err := s.db.ExecContext(
		ctx,
		query,
		userToFollowId,
		userId,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
	}

	return err
}

func (s *UsersRepositoryPostgres) Unfollow(ctx context.Context, userToUnfollowId int64, userId int64) error {
	query := `
	DELETE FROM followers WHERE (user_id, follower_id) = ($1, $2)  
	`
	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()
	// Execute the query and scan the result into the `post` struct.
	res, err := s.db.ExecContext(ctx, query, userToUnfollowId, userId)
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

func (s *UsersRepositoryPostgres) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// get correct user
		user, err := s.getUserFromInvite(ctx, tx, token)
		if err != nil {
			return err
		}

		//  update user
		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		//  clean invitation table
		if err := s.deleteInvite(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UsersRepositoryPostgres) getUserFromInvite(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.created_at, u.is_active
		FROM users u JOIN user_invitations ui ON u.id = ui.user_id 
		WHERE ui.token = $1 and ui.expiry > $2
	`

	var user User

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UsersRepositoryPostgres) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
	UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4
	`

	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UsersRepositoryPostgres) deleteInvite(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `
	DELETE FROM user_invitations WHERE user_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersRepositoryPostgres) deleteUser(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
		DELETE FROM users WHERE id = $1 
	`
	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ErrNotFound
	default:
		return err
	}
}

func (s *UsersRepositoryPostgres) Delete(ctx context.Context, id int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.deleteUser(ctx, tx, id); err != nil {
			return err
		}

		if err := s.deleteInvite(ctx, tx, id); err != nil {
			return err
		}

		return nil
	})
}

func (s *UsersRepositoryPostgres) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, username, password, created_at FROM users WHERE email = $1 AND is_active = true
	`
	var user User
	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.Password.hash,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
