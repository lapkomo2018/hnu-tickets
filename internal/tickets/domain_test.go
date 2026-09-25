package tickets

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTicketChangeStatus(t *testing.T) {
	now := time.Date(2026, time.September, 25, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		from Status
		to   Status
	}{
		{name: "assign new ticket", from: StatusNew, to: StatusAssigned},
		{name: "start assigned ticket", from: StatusAssigned, to: StatusInProgress},
		{name: "wait for user", from: StatusInProgress, to: StatusWaitingForUser},
		{name: "resume after user response", from: StatusWaitingForUser, to: StatusInProgress},
		{name: "resolve ticket", from: StatusInProgress, to: StatusResolved},
		{name: "reopen ticket", from: StatusResolved, to: StatusInProgress},
		{name: "close ticket", from: StatusResolved, to: StatusClosed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := Ticket{Status: tt.from}
			require.NoError(t, ticket.ChangeStatus(tt.to, now))
			require.Equal(t, tt.to, ticket.Status)
			require.Equal(t, now, ticket.UpdatedAt)
		})
	}
}

func TestTicketChangeStatusRejectsInvalidTransition(t *testing.T) {
	ticket := Ticket{Status: StatusNew}
	err := ticket.ChangeStatus(StatusClosed, time.Now())
	require.ErrorIs(t, err, ErrInvalidTransition)
	require.Equal(t, StatusNew, ticket.Status)
}

func TestTicketChangeStatusUpdatesLifecycleTimestamps(t *testing.T) {
	resolvedAt := time.Date(2026, time.September, 25, 12, 0, 0, 0, time.UTC)
	ticket := Ticket{Status: StatusInProgress}

	require.NoError(t, ticket.ChangeStatus(StatusResolved, resolvedAt))
	require.Equal(t, &resolvedAt, ticket.ResolvedAt)

	require.NoError(t, ticket.ChangeStatus(StatusInProgress, resolvedAt.Add(time.Hour)))
	require.Nil(t, ticket.ResolvedAt)

	require.NoError(t, ticket.ChangeStatus(StatusResolved, resolvedAt.Add(2*time.Hour)))
	closedAt := resolvedAt.Add(3 * time.Hour)
	require.NoError(t, ticket.ChangeStatus(StatusClosed, closedAt))
	require.Equal(t, &closedAt, ticket.ClosedAt)
}
