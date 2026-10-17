package tickets

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/users"
)

func TestTicketVisibilityPolicy(t *testing.T) {
	creatorID := uuid.New()
	specialistID := uuid.New()
	ticket := &Ticket{CreatedBy: creatorID, AssignedTo: &specialistID}

	require.True(t, canView(appmiddleware.Principal{UserID: creatorID, Role: string(users.RoleUser)}, ticket))
	require.True(t, canView(appmiddleware.Principal{UserID: specialistID, Role: string(users.RoleSpecialist)}, ticket))
	require.True(t, canView(appmiddleware.Principal{UserID: uuid.New(), Role: string(users.RoleAdmin)}, ticket))
	require.False(t, canView(appmiddleware.Principal{UserID: uuid.New(), Role: string(users.RoleUser)}, ticket))
	require.False(t, canView(appmiddleware.Principal{UserID: uuid.New(), Role: string(users.RoleSpecialist)}, ticket))
}

func TestAuthorEditRequiresCurrentUserRole(t *testing.T) {
	id := uuid.New()
	ticket := &Ticket{CreatedBy: id, Status: StatusNew}
	require.True(t, canEdit(appmiddleware.Principal{UserID: id, Role: "USER"}, ticket))
	require.False(t, canEdit(appmiddleware.Principal{UserID: id, Role: "SPECIALIST"}, ticket))
	require.False(t, canEdit(appmiddleware.Principal{UserID: uuid.New(), Role: "USER"}, ticket))
	ticket.Status = StatusAssigned
	require.False(t, canEdit(appmiddleware.Principal{UserID: id, Role: "USER"}, ticket))
	require.True(t, canEdit(appmiddleware.Principal{UserID: id, Role: "ADMIN"}, ticket))
}

func TestStatusPolicyRestrictsActors(t *testing.T) {
	creatorID := uuid.New()
	specialistID := uuid.New()
	ticket := &Ticket{CreatedBy: creatorID, AssignedTo: &specialistID, Status: StatusAssigned}

	require.NoError(t, authorizeStatusChange(
		appmiddleware.Principal{UserID: specialistID, Role: string(users.RoleSpecialist)}, ticket, StatusInProgress,
	))
	require.Error(t, authorizeStatusChange(
		appmiddleware.Principal{UserID: uuid.New(), Role: string(users.RoleSpecialist)}, ticket, StatusInProgress,
	))
	require.Error(t, authorizeStatusChange(
		appmiddleware.Principal{UserID: creatorID, Role: string(users.RoleUser)}, ticket, StatusInProgress,
	))

	ticket.Status = StatusResolved
	require.NoError(t, authorizeStatusChange(
		appmiddleware.Principal{UserID: creatorID, Role: string(users.RoleUser)}, ticket, StatusClosed,
	))
}
