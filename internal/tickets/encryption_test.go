package tickets

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"hnu-tickets/internal/platform/fieldcrypt"
)

type encryptedRow struct {
	id          uuid.UUID
	description string
}

func (row encryptedRow) Scan(dest ...any) error {
	*dest[0].(*uuid.UUID) = row.id
	*dest[2].(*string) = row.description
	return nil
}

func TestTicketReadersDecryptAndRejectAnotherRowsCiphertext(t *testing.T) {
	keys, err := fieldcrypt.New("test", `{"test":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}`)
	require.NoError(t, err)
	id := uuid.New()
	sealed, err := keys.Encrypt("Конфіденційний опис", fieldcrypt.Context("tickets", id.String(), "description"))
	require.NoError(t, err)
	r := NewPostgresRepository(nil, keys)
	row := encryptedRow{id, sealed}
	ticket, err := r.scanTicket(row)
	require.NoError(t, err)
	require.Equal(t, "Конфіденційний опис", ticket.Description)
	record, err := r.scanRecord(row)
	require.NoError(t, err)
	require.Equal(t, ticket.Description, record.Description)
	row.id = uuid.New()
	_, err = r.scanTicket(row)
	require.ErrorIs(t, err, fieldcrypt.ErrInvalid)
	_, err = r.scanRecord(row)
	require.ErrorIs(t, err, fieldcrypt.ErrInvalid)
	row.description = "plaintext"
	_, err = r.scanTicket(row)
	require.ErrorIs(t, err, fieldcrypt.ErrInvalid)
}
