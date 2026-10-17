package fieldcrypt

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const testKeys = `{"old":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","new":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="}`

func TestEnvelopeProtectsContentAndContext(t *testing.T) {
	r, err := New("old", testKeys)
	require.NoError(t, err)
	context := Context("tickets", "id-1", "description")
	for _, plain := range []string{"", "Опис заявки 🔐", "enc:v1:user supplied text"} {
		sealed, err := r.Encrypt(plain, context)
		require.NoError(t, err)
		again, err := r.Encrypt(plain, context)
		require.NoError(t, err)
		require.NotEqual(t, sealed, again, "each write needs a fresh nonce")
		decoded, err := r.Decrypt(sealed, context)
		require.NoError(t, err)
		require.Equal(t, plain, decoded)
		for _, wrong := range []string{Context("tickets", "id-2", "description"), Context("tickets", "id-1", "title"), Context("comments", "id-1", "description")} {
			decoded, err := r.Decrypt(sealed, wrong)
			require.ErrorIs(t, err, ErrInvalid)
			require.Empty(t, decoded)
		}
		parts := strings.Split(sealed, ":")
		data, err := base64.RawStdEncoding.DecodeString(parts[3])
		require.NoError(t, err)
		for _, offset := range []int{0, 12, len(data) - 1} {
			changed := append([]byte(nil), data...)
			changed[offset] ^= 1
			decoded, err := r.Decrypt(strings.Join(parts[:3], ":")+":"+base64.RawStdEncoding.EncodeToString(changed), context)
			require.ErrorIs(t, err, ErrInvalid)
			require.Empty(t, decoded)
		}
	}
}

func TestRotationAndInvalidEnvelopes(t *testing.T) {
	old, err := New("old", testKeys)
	require.NoError(t, err)
	current, err := New("new", testKeys)
	require.NoError(t, err)
	sealed, err := old.Encrypt("original", "context")
	require.NoError(t, err)
	plain, err := current.Decrypt(sealed, "context")
	require.NoError(t, err)
	rotated, err := current.Encrypt(plain, "context")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(rotated, "enc:v1:new:"))
	onlyNew, err := New("new", `{"new":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="}`)
	require.NoError(t, err)
	plain, err = onlyNew.Decrypt(rotated, "context")
	require.NoError(t, err)
	require.Equal(t, "original", plain)
	for _, invalid := range []string{sealed, "plaintext", "enc:v2:new:AA", "enc:v1:new:", "enc:v1:new:???", strings.Replace(rotated, ":new:", ":old:", 1)} {
		plain, err := onlyNew.Decrypt(invalid, "context")
		require.ErrorIs(t, err, ErrInvalid)
		require.Empty(t, plain)
	}
	wrong, err := New("new", `{"new":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI="}`)
	require.NoError(t, err)
	_, err = wrong.Decrypt(rotated, "context")
	require.ErrorIs(t, err, ErrInvalid)
}

func TestRejectsInvalidConfigurationWithoutEchoingSecrets(t *testing.T) {
	for _, input := range []struct{ active, keys string }{
		{"old", ""}, {"old", "{}"}, {"absent", testKeys}, {"old", `{"old":"TOP-SECRET"}`},
		{"bad:id", `{"bad:id":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}`},
	} {
		_, err := New(input.active, input.keys)
		require.Error(t, err)
		require.NotContains(t, err.Error(), "TOP-SECRET")
	}
}
