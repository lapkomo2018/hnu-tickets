package history

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/platform/fieldcrypt"
)

func TestHistoryPreservesNullAndSeparatesOldAndNewValues(t *testing.T) {
	keys, err := fieldcrypt.New("test", `{"test":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}`)
	require.NoError(t, err)
	r := NewPostgresRepository(nil, keys)
	id := uuid.New()
	value, err := r.encryptValue(id, "old_value", nil)
	require.NoError(t, err)
	require.Nil(t, value)
	text := ""
	value, err = r.encryptValue(id, "old_value", &text)
	require.NoError(t, err)
	require.NotNil(t, value)
	plain, err := keys.Decrypt(*value, fieldcrypt.Context("ticket_history", id.String(), "old_value"))
	require.NoError(t, err)
	require.Empty(t, plain)
	_, err = keys.Decrypt(*value, fieldcrypt.Context("ticket_history", id.String(), "new_value"))
	require.ErrorIs(t, err, fieldcrypt.ErrInvalid)
}
