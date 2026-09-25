package analytics

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/users"
)

func TestAuthorizeRequiresAdmin(t *testing.T) {
	err := authorize(appmiddleware.Principal{UserID: uuid.New(), Role: string(users.RoleUser)}, Period{})
	require.Error(t, err)

	err = authorize(appmiddleware.Principal{UserID: uuid.New(), Role: string(users.RoleAdmin)}, Period{})
	require.NoError(t, err)
}

func TestAuthorizeRejectsReversedPeriod(t *testing.T) {
	from := time.Date(2026, time.September, 26, 0, 0, 0, 0, time.UTC)
	to := from.Add(-time.Hour)
	err := authorize(appmiddleware.Principal{UserID: uuid.New(), Role: string(users.RoleAdmin)}, Period{From: &from, To: &to})
	require.Error(t, err)
}
