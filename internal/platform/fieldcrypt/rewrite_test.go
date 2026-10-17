package fieldcrypt

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// These fakes exercise the offline rewrite without opening a database.
type rewriteFake struct {
	pgx.Tx
	encrypted      bool
	values         map[string][]storedValue
	reads, updates int
}

type storedValue struct{ id, text string }

func (f *rewriteFake) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.HasPrefix(sql, "UPDATE data_crypto_state") {
		f.encrypted = true
	}
	for _, field := range protectedFields {
		if strings.HasPrefix(sql, `UPDATE "`+field.table+`" SET "`+field.column+`"`) {
			key := field.table + "." + field.column
			for i, v := range f.values[key] {
				if v.id == args[1] {
					f.values[key][i].text = args[0].(string)
					f.updates++
				}
			}
		}
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

type boolRow bool

func (r boolRow) Scan(dest ...any) error { *dest[0].(*bool) = bool(r); return nil }

func (f *rewriteFake) QueryRow(context.Context, string, ...any) pgx.Row { return boolRow(f.encrypted) }

func (f *rewriteFake) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	f.reads++
	for _, field := range protectedFields {
		if strings.Contains(sql, `FROM "`+field.table+`"`) && strings.Contains(sql, `, "`+field.column+`" FROM`) {
			batch := []storedValue{}
			for _, v := range f.values[field.table+"."+field.column] {
				if args[0] == nil || v.id > args[0].(string) {
					batch = append(batch, v)
					if len(batch) == 500 {
						break
					}
				}
			}
			return &fakeRows{values: batch}, nil
		}
	}
	return nil, fmt.Errorf("unexpected query")
}

type fakeRows struct {
	pgx.Rows
	values []storedValue
	index  int
}

func (r *fakeRows) Next() bool { r.index++; return r.index <= len(r.values) }
func (r *fakeRows) Scan(dest ...any) error {
	v := r.values[r.index-1]
	*dest[0].(*string) = v.id
	*dest[1].(*string) = v.text
	return nil
}
func (r *fakeRows) Err() error { return nil }
func (r *fakeRows) Close()     {}

func TestRewriteMigratesAllFieldsAndRotatesInBatches(t *testing.T) {
	old, err := New("old", testKeys)
	require.NoError(t, err)
	current, err := New("new", testKeys)
	require.NoError(t, err)
	f := &rewriteFake{values: map[string][]storedValue{}}
	for _, field := range protectedFields {
		for i := 0; i < 501; i++ {
			f.values[field.table+"."+field.column] = append(f.values[field.table+"."+field.column], storedValue{fmt.Sprintf("00000000-0000-0000-0000-%012d", i), "enc:v1:literal user text"})
		}
	}
	n, err := rewriteTransaction(t.Context(), f, old, "migrate")
	require.NoError(t, err)
	require.EqualValues(t, 2004, n)
	require.True(t, f.encrypted)
	n, err = rewriteTransaction(t.Context(), f, current, "rotate")
	require.NoError(t, err)
	require.EqualValues(t, 2004, n)
	for _, field := range protectedFields {
		for _, v := range f.values[field.table+"."+field.column] {
			require.True(t, strings.HasPrefix(v.text, "enc:v1:new:"))
			plain, err := current.Decrypt(v.text, Context(field.table, v.id, field.column))
			require.NoError(t, err)
			require.Equal(t, "enc:v1:literal user text", plain)
		}
	}
	_, err = rewriteTransaction(t.Context(), f, current, "migrate")
	require.Error(t, err)
}

func TestRewriteRejectsCorruptInputAndWrongMode(t *testing.T) {
	keys, err := New("new", testKeys)
	require.NoError(t, err)
	f := &rewriteFake{encrypted: true, values: map[string][]storedValue{"tickets.description": {{"00000000-0000-0000-0000-000000000001", "damaged"}}}}
	_, err = rewriteTransaction(t.Context(), f, keys, "rotate")
	require.ErrorIs(t, err, ErrInvalid)
	require.Zero(t, f.updates)
	f.encrypted = false
	_, err = rewriteTransaction(t.Context(), f, keys, "rotate")
	require.Error(t, err)
	require.Zero(t, f.updates)
}
