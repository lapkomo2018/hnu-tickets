package auth

import (
	"context"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/users"
)

type memoryUsers struct {
	users.Repository
	byEmail map[string]*users.User
}

func newMemoryUsers(existing ...users.User) *memoryUsers {
	repository := &memoryUsers{byEmail: map[string]*users.User{}}
	for _, user := range existing {
		repository.byEmail[user.Email] = &user
	}
	return repository
}

func (r *memoryUsers) GetByEmail(_ context.Context, email string) (*users.User, error) {
	user, ok := r.byEmail[email]
	if !ok {
		return nil, users.ErrUserNotFound
	}
	return user, nil
}

func (r *memoryUsers) Create(_ context.Context, user *users.User) error {
	if _, ok := r.byEmail[user.Email]; ok {
		return users.ErrEmailAlreadyExists
	}
	user.ID = uuid.New()
	r.byEmail[user.Email] = user
	return nil
}

func TestBootstrapAdminCreatesAdministrator(t *testing.T) {
	repository := newMemoryUsers()

	err := BootstrapAdmin(t.Context(), repository, RegisterInput{
		Name: "Administrator", Email: " Admin@Example.com ", Password: "long-password",
	})

	require.NoError(t, err)
	user := repository.byEmail["admin@example.com"]
	require.NotNil(t, user)
	require.Equal(t, users.RoleAdmin, user.Role)
	require.NoError(t, checkPassword(user.PasswordHash, "long-password"))
}

func TestBootstrapAdminLeavesExistingUserUntouched(t *testing.T) {
	repository := newMemoryUsers(users.User{Email: "admin@example.com", PasswordHash: "original", Role: users.RoleUser})

	err := BootstrapAdmin(t.Context(), repository, RegisterInput{
		Name: "Administrator", Email: "admin@example.com", Password: "long-password",
	})

	require.NoError(t, err)
	user := repository.byEmail["admin@example.com"]
	require.Equal(t, users.RoleUser, user.Role)
	require.Equal(t, "original", user.PasswordHash)
}

func TestBootstrapAdminRejectsWeakPassword(t *testing.T) {
	err := BootstrapAdmin(t.Context(), newMemoryUsers(), RegisterInput{
		Name: "Administrator", Email: "admin@example.com", Password: "short",
	})

	require.ErrorIs(t, err, ErrInvalidPassword)
}
