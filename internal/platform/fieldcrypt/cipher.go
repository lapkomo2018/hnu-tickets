// Package fieldcrypt protects stored text with versioned AES-256-GCM envelopes.
package fieldcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrInvalid = errors.New("encrypted field authentication failed")

var keyID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,32}$`)

type Keyring struct {
	active string
	keys   map[string]cipher.AEAD
}

// New accepts a JSON object mapping key IDs to base64-encoded 32-byte keys.
// Key material is never included in errors.
func New(active, encoded string) (*Keyring, error) {
	var values map[string]string
	if err := json.Unmarshal([]byte(encoded), &values); err != nil || len(values) == 0 {
		return nil, errors.New("DATA_KEYS must be a nonempty JSON object")
	}
	r := &Keyring{active: active, keys: make(map[string]cipher.AEAD)}
	for id, value := range values {
		if !keyID.MatchString(id) {
			return nil, errors.New("invalid data key ID")
		}
		key, err := base64.StdEncoding.Strict().DecodeString(value)
		if err != nil || len(key) != 32 {
			return nil, errors.New("data keys must contain 32 base64-encoded bytes")
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, errors.New("initialize AES")
		}
		aead, err := cipher.NewGCMWithRandomNonce(block)
		if err != nil {
			return nil, errors.New("initialize GCM")
		}
		r.keys[id] = aead
	}
	if r.keys[active] == nil {
		return nil, errors.New("DATA_ACTIVE_KEY must identify a configured key")
	}
	return r, nil
}

func (r *Keyring) Encrypt(plaintext, context string) (string, error) {
	if r == nil || r.keys[r.active] == nil {
		return "", errors.New("data encryption is not configured")
	}
	header := "enc:v1:" + r.active + ":"
	// The standard library generates and prepends a fresh 96-bit nonce.
	sealed := r.keys[r.active].Seal(nil, nil, []byte(plaintext), []byte(header+context))
	return header + base64.RawStdEncoding.EncodeToString(sealed), nil
}

func (r *Keyring) Decrypt(envelope, context string) (string, error) {
	parts := strings.Split(envelope, ":")
	if r == nil || len(parts) != 4 || parts[0] != "enc" || parts[1] != "v1" {
		return "", ErrInvalid
	}
	aead := r.keys[parts[2]]
	if aead == nil {
		return "", ErrInvalid
	}
	sealed, err := base64.RawStdEncoding.Strict().DecodeString(parts[3])
	if err != nil {
		return "", ErrInvalid
	}
	plain, err := aead.Open(nil, nil, sealed, []byte(strings.Join(parts[:3], ":")+":"+context))
	if err != nil {
		return "", ErrInvalid
	}
	return string(plain), nil
}

func Context(table, id, column string) string {
	return fmt.Sprintf("hnu-tickets/%s/%s/%s", table, id, column)
}
