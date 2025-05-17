package store

import (
	"context"
	"database/sql"
	"errors"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type RoleRepositoryPostgres struct {
	db *sql.DB
}

func (s *RoleRepositoryPostgres) GetByName(ctx context.Context, roleName string) (*Role, error) {
	query := `
		SELECT 	id, name, level, description FROM Roles WHERE name = $1
	`

	var role Role
	ctx, cancel := context.WithTimeout(ctx, QueryContextTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, roleName).Scan(
		&role.ID,
		&role.Name,
		&role.Level,
		&role.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &role, nil
}
