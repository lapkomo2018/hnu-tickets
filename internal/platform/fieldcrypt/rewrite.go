package fieldcrypt

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RequireReady(ctx context.Context, pool *pgxpool.Pool) error {
	var ready bool
	if err := pool.QueryRow(ctx, "SELECT encrypted FROM data_crypto_state WHERE id = TRUE").Scan(&ready); err != nil {
		return fmt.Errorf("read encryption state: %w", err)
	}
	if !ready {
		return errors.New("run cryptodata -mode migrate before starting the API")
	}
	return nil
}

type field struct{ table, column string }

var protectedFields = []field{
	{"tickets", "description"}, {"comments", "content"},
	{"ticket_history", "old_value"}, {"ticket_history", "new_value"},
}

// Rewrite runs offline under exclusive table locks. It never guesses whether a
// value is plaintext from its prefix: the database state selects the input form.
// All fields and the state marker commit together or roll back together.
func Rewrite(ctx context.Context, pool *pgxpool.Pool, keys *Keyring, mode string) (int64, error) {
	if mode != "migrate" && mode != "rotate" {
		return 0, errors.New("mode must be migrate or rotate")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	count, err := rewriteTransaction(ctx, tx, keys, mode)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return count, nil
}

func rewriteTransaction(ctx context.Context, tx pgx.Tx, keys *Keyring, mode string) (int64, error) {
	_, err := tx.Exec(ctx, "LOCK TABLE data_crypto_state, tickets, comments, ticket_history IN ACCESS EXCLUSIVE MODE")
	if err != nil {
		return 0, err
	}
	var encrypted bool
	if err := tx.QueryRow(ctx, "SELECT encrypted FROM data_crypto_state WHERE id = TRUE").Scan(&encrypted); err != nil {
		return 0, err
	}
	if (mode == "rotate") != encrypted {
		return 0, errors.New("mode does not match database encryption state")
	}
	var count int64
	for _, f := range protectedFields {
		n, err := rewriteField(ctx, tx, keys, f, encrypted)
		if err != nil {
			return 0, fmt.Errorf("rewrite %s.%s: %w", f.table, f.column, err)
		}
		count += n
	}
	if _, err := tx.Exec(ctx, "UPDATE data_crypto_state SET encrypted = TRUE WHERE id = TRUE"); err != nil {
		return 0, err
	}
	return count, nil
}

func rewriteField(ctx context.Context, tx pgx.Tx, keys *Keyring, f field, encrypted bool) (int64, error) {
	// Names come only from the fixed protectedFields list, never user input.
	table, column := pgx.Identifier{f.table}.Sanitize(), pgx.Identifier{f.column}.Sanitize()
	var count int64
	var last any
	for {
		rows, err := tx.Query(ctx, "SELECT id::text, "+column+" FROM "+table+" WHERE ($1::uuid IS NULL OR id > $1::uuid) AND "+column+" IS NOT NULL ORDER BY id LIMIT 500", last)
		if err != nil {
			return 0, err
		}
		type value struct{ id, text string }
		batch := make([]value, 0, 500)
		for rows.Next() {
			var v value
			if err := rows.Scan(&v.id, &v.text); err != nil {
				rows.Close()
				return 0, err
			}
			batch = append(batch, v)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return 0, err
		}
		if len(batch) == 0 {
			return count, nil
		}
		for _, v := range batch {
			context := Context(f.table, v.id, f.column)
			plain := v.text
			if encrypted {
				plain, err = keys.Decrypt(v.text, context)
				if err != nil {
					return 0, err
				}
			}
			sealed, err := keys.Encrypt(plain, context)
			if err != nil {
				return 0, err
			}
			if _, err := tx.Exec(ctx, "UPDATE "+table+" SET "+column+" = $1 WHERE id = $2::uuid", sealed, v.id); err != nil {
				return 0, err
			}
			last = v.id
			count++
		}
	}
}
