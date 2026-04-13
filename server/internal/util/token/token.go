// Package token produces URL-safe, base64-encoded random tokens.
// Shared by the auth package (session credentials) and the lobby
// package (game invite tokens) so the two stay in lockstep on
// entropy + encoding choices.
package token

import (
	"crypto/rand"
	"encoding/base64"
)

// Random returns n bytes of crypto-random data encoded as
// base64url (URL-safe, padding-free). The result is 4*⌈n/3⌉ bytes
// long. Panics are not used; I/O errors from crypto/rand (which are
// only possible on a broken host) bubble up.
func Random(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
