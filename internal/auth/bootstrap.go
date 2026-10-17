package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"hnu-tickets/internal/users"
)

// BootstrapAdmin creates the first administrator so a fresh database is
// usable without seed data. It runs before the server accepts requests and
// only ever creates an account: an existing user with the same email is left
// untouched, so changing the configured email cannot promote someone else's
// account.
func BootstrapAdmin(ctx context.Context, repository users.Repository, input RegisterInput) error {
	name, email, err := validateCredentials(input.Name, input.Email, input.Password)
	if err != nil {
		return fmt.Errorf("validate bootstrap admin: %w", err)
	}

	existing, err := repository.GetByEmail(ctx, email)
	switch {
	case err == nil:
		if existing.Role != users.RoleAdmin {
			slog.Warn("bootstrap admin email belongs to a non-admin user, skipping", "email", email, "role", existing.Role)
		}
		return nil
	case !errors.Is(err, users.ErrUserNotFound):
		return fmt.Errorf("look up bootstrap admin: %w", err)
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user := &users.User{Name: name, Email: email, PasswordHash: hash, Role: users.RoleAdmin}
	if err := repository.Create(ctx, user); err != nil {
		// Another replica created the account between the lookup and the insert.
		if errors.Is(err, users.ErrEmailAlreadyExists) {
			return nil
		}
		return fmt.Errorf("create bootstrap admin: %w", err)
	}
	slog.Info("bootstrap admin created", "email", email)
	return nil
}
