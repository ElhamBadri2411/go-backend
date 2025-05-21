package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStore() Storage {
	return Storage{
		UsersRepository: &MockUserStore{},
	}
}

type MockUserStore struct{}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	return nil
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitation_expiry time.Duration) error {
	return nil
}

func (m *MockUserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, userId int64, invitation_expiry time.Duration) error {
	return nil
}

func (m *MockUserStore) GetById(ctx context.Context, id int64) (*User, error) {
	return &User{ID: id}, nil
}

func (m *MockUserStore) Follow(ctx context.Context, userToFollowId int64, userId int64) error {
	return nil
}

func (m *MockUserStore) Unfollow(ctx context.Context, userToUnfollowId int64, userId int64) error {
	return nil
}

func (m *MockUserStore) Activate(ctx context.Context, token string) error {
	return nil
}

func (m *MockUserStore) getUserFromInvite(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	return &User{}, nil
}

func (m *MockUserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	return nil
}

func (m *MockUserStore) deleteInvite(ctx context.Context, tx *sql.Tx, userId int64) error {
	return nil
}

func (m *MockUserStore) deleteUser(ctx context.Context, tx *sql.Tx, id int64) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return &User{}, nil
}
