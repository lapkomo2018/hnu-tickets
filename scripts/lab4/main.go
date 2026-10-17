// Lab 4 demonstrates classical ciphers and CTR ciphertext malleability.
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func caesar(text string, shift int) string {
	alphabets := []string{"абвгґдеєжзиіїйклмнопрстуфхцчшщьюя", "АБВГҐДЕЄЖЗИІЇЙКЛМНОПРСТУФХЦЧШЩЬЮЯ", "abcdefghijklmnopqrstuvwxyz", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"}
	return strings.Map(func(r rune) rune {
		for _, alphabet := range alphabets {
			a := []rune(alphabet)
			for i, c := range a {
				if r == c {
					return a[(i+shift+len(a))%len(a)]
				}
			}
		}
		return r
	}, text)
}

var order = []int{2, 0, 6, 1, 4, 3, 5}

func transpose(text string) string {
	runes, out := []rune(text), []rune{}
	for _, column := range order {
		for i := column; i < len(runes); i += len(order) {
			out = append(out, runes[i])
		}
	}
	return string(out)
}

func untranspose(text string) string {
	runes, out := []rune(text), make([]rune, len([]rune(text)))
	n := 0
	for _, column := range order {
		for i := column; i < len(out); i += len(order) {
			out[i] = runes[n]
			n++
		}
	}
	return string(out)
}

func xor(a, b []byte) []byte {
	if len(a) != len(b) {
		panic("XOR requires equal byte lengths")
	}
	out := make([]byte, len(a))
	for i := range a {
		out[i] = a[i] ^ b[i]
	}
	return out
}

func main() {
	output := flag.String("output", "scripts/lab4/results.json", "result JSON path")
	flag.Parse()
	if err := run(*output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(output string) error {
	plain := "2026_ІПЗс-24-1_Лапко_Максим_варіант:07_"
	alternative := strings.Replace(plain, ":07_", ":99_", 1)
	key, iv := make([]byte, 32), make([]byte, aes.BlockSize)
	if _, err := rand.Read(key); err != nil {
		return err
	}
	if _, err := rand.Read(iv); err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	first := make([]byte, len([]byte(plain)))
	cipher.NewCTR(block, iv).XORKeyStream(first, []byte(plain))
	// The attacker receives only the known pair; no key or block cipher.
	gamma := xor(first, []byte(plain))
	forged := xor(gamma, []byte(alternative))
	recovered := make([]byte, len(forged))
	cipher.NewCTR(block, iv).XORKeyStream(recovered, forged)
	if string(recovered) != alternative || caesar(caesar(plain, 7), -7) != plain || untranspose(transpose(plain)) != plain {
		return fmt.Errorf("experiment verification failed")
	}
	result := map[string]any{
		"created_utc": time.Now().UTC().Format(time.RFC3339), "plaintext": plain, "utf8_hex": hex.EncodeToString([]byte(plain)),
		"characters": len([]rune(plain)), "utf8_bytes": len([]byte(plain)),
		"caesar_shift": 7, "caesar_ciphertext": caesar(plain, 7), "caesar_roundtrip": true,
		"transposition_columns_zero_based": order, "transposition_ciphertext": transpose(plain), "transposition_roundtrip": true,
		"algorithm": "AES-256-CTR", "counter": "128-bit big-endian counter initialized to IV, incremented per block", "padding": "none",
		"key_1_hex": hex.EncodeToString(key), "iv_1_hex": hex.EncodeToString(iv), "ciphertext_1_hex": hex.EncodeToString(first),
		"gamma_1_hex": hex.EncodeToString(gamma), "alternative_plaintext": alternative, "ciphertext_2_hex": hex.EncodeToString(forged),
		"decrypted_ciphertext_2": string(recovered), "attack_verified": true,
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(output, append(data, '\n'), 0o644)
}
