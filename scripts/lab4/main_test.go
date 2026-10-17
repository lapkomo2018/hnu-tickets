package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecordedExperiment(t *testing.T) {
	data, err := os.ReadFile("results.json")
	require.NoError(t, err)
	var r map[string]any
	require.NoError(t, json.Unmarshal(data, &r))
	get := func(key string) string { return r[key].(string) }
	decode := func(key string) []byte { b, err := hex.DecodeString(get(key)); require.NoError(t, err); return b }
	plain := get("plaintext")
	require.Equal(t, get("caesar_ciphertext"), caesar(plain, 7))
	require.Equal(t, plain, caesar(caesar(plain, 7), -7))
	require.Equal(t, get("transposition_ciphertext"), transpose(plain))
	require.Equal(t, plain, untranspose(transpose(plain)))
	block, err := aes.NewCipher(decode("key_1_hex"))
	require.NoError(t, err)
	first := make([]byte, len([]byte(plain)))
	cipher.NewCTR(block, decode("iv_1_hex")).XORKeyStream(first, []byte(plain))
	require.Equal(t, decode("ciphertext_1_hex"), first)
	gamma := xor(first, []byte(plain))
	require.Equal(t, decode("gamma_1_hex"), gamma)
	forged := xor(gamma, []byte(get("alternative_plaintext")))
	require.Equal(t, decode("ciphertext_2_hex"), forged)
	decoded := make([]byte, len(forged))
	cipher.NewCTR(block, decode("iv_1_hex")).XORKeyStream(decoded, forged)
	require.Equal(t, get("alternative_plaintext"), string(decoded))
}
