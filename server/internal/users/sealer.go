package users

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"strings"
)

// IdentityKeyEnv names the environment variable holding the key that
// encrypts OAuth refresh tokens at rest (ADR 0051 decision 5). It
// lives in /etc/cmd_and_ctrl/env beside CMDCTRL_ADMIN_TOKEN and
// CMDCTRL_SESSION_KEY.
const IdentityKeyEnv = "CMDCTRL_IDENTITY_KEY"

// MinIdentityKeyBytes is the shortest key NewSealer accepts: the same
// floor, and the same format — a random string, not decoded — as
// CMDCTRL_SESSION_KEY, so one generator line serves both.
const MinIdentityKeyBytes = 32

// sealVersion1 is the first byte of every sealed value. It names the
// key and construction that sealed it, so a rotation can add a
// version 2 key and still open (then re-seal) version 1 rows.
const sealVersion1 byte = 1

// keyDerivationLabel separates this key's derived AES key from any
// other use of the same bytes. The env value is operator-supplied
// text of any length >= 32; SHA-256 over label||value turns it into
// exactly the 32 bytes AES-256 needs.
const keyDerivationLabel = "cmd_and_ctrl identity key v1\x00"

var (
	// ErrIdentityKeyTooShort: the key is set but under
	// MinIdentityKeyBytes.
	ErrIdentityKeyTooShort = fmt.Errorf("users: %s must be at least %d bytes", IdentityKeyEnv, MinIdentityKeyBytes)
	// ErrIdentityKeyReused: the key equals the admin token or the
	// session key. Each secret has one job, so that leaking one (the
	// admin token is copied into the bot's env file) does not also
	// unlock stored refresh tokens.
	ErrIdentityKeyReused = fmt.Errorf("users: %s must differ from CMDCTRL_ADMIN_TOKEN and CMDCTRL_SESSION_KEY", IdentityKeyEnv)
	// ErrSealed: a sealed value failed to open — truncated, tampered
	// with, sealed under another key, or an unknown version.
	ErrSealed = errors.New("users: sealed value cannot be opened")
)

// Sealer encrypts and authenticates small secrets with AES-256-GCM.
//
// # Sealed format
//
//	version (1 byte) || nonce (12 bytes) || ciphertext || tag (16 bytes)
//
// The nonce is fresh from crypto/rand on every Seal. The caller's
// associated data (for a refresh token: provider and subject) is bound
// into the tag but not stored, so a sealed blob copied onto another
// identity's row does not open there.
type Sealer struct {
	aead cipher.AEAD
}

// NewSealer derives an AES-256 key from rawKey (see
// keyDerivationLabel). A key under MinIdentityKeyBytes is refused.
func NewSealer(rawKey string) (*Sealer, error) {
	if len(rawKey) < MinIdentityKeyBytes {
		return nil, ErrIdentityKeyTooShort
	}
	sum := sha256.Sum256([]byte(keyDerivationLabel + rawKey))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Sealer{aead: aead}, nil
}

// Seal encrypts plaintext, binding ad into the tag.
func (s *Sealer) Seal(plaintext, ad []byte) ([]byte, error) {
	ns := s.aead.NonceSize()
	out := make([]byte, 1+ns, 1+ns+len(plaintext)+s.aead.Overhead())
	out[0] = sealVersion1
	if _, err := io.ReadFull(rand.Reader, out[1:1+ns]); err != nil {
		return nil, fmt.Errorf("users: nonce: %w", err)
	}
	return s.aead.Seal(out, out[1:1+ns], plaintext, ad), nil
}

// Open reverses Seal. Any failure is ErrSealed; the reason is not
// distinguished, so a caller cannot be used as an oracle.
func (s *Sealer) Open(sealed, ad []byte) ([]byte, error) {
	ns := s.aead.NonceSize()
	if len(sealed) < 1+ns+s.aead.Overhead() || sealed[0] != sealVersion1 {
		return nil, ErrSealed
	}
	pt, err := s.aead.Open(nil, sealed[1:1+ns], sealed[1+ns:], ad)
	if err != nil {
		return nil, ErrSealed
	}
	return pt, nil
}

// Logger is the one method NewSealerFromEnv needs. *slog.Logger
// satisfies it.
type Logger interface {
	Warn(msg string, args ...any)
}

// NewSealerFromEnv reads IdentityKeyEnv through getenv:
//
//   - Unset or blank: (nil, nil), with a warning naming the variable.
//     Sign-in still works; refresh tokens are discarded at the
//     callback and stored as NULL. A missing key never fails open.
//   - Set but shorter than MinIdentityKeyBytes, or equal to
//     CMDCTRL_ADMIN_TOKEN or CMDCTRL_SESSION_KEY: an error. The
//     caller (main) exits rather than boot on a weak or shared key.
//   - Otherwise: a Sealer.
//
// Whitespace around each value is trimmed, as NewFromEnv does for the
// session key, so the comparison sees what the servers actually use.
func NewSealerFromEnv(getenv func(string) string, log Logger) (*Sealer, error) {
	raw := strings.TrimSpace(getenv(IdentityKeyEnv))
	if raw == "" {
		log.Warn(IdentityKeyEnv+" is not set; Discord refresh tokens are DISCARDED at sign-in and stored as NULL. Sign-in still works. Set it to a random string of at least 32 bytes (e.g. `head -c 48 /dev/urandom | base64 -w0`), different from the admin token and the session key.",
			"var", IdentityKeyEnv)
		return nil, nil
	}
	if len(raw) < MinIdentityKeyBytes {
		return nil, fmt.Errorf("%s is %d bytes: %w", IdentityKeyEnv, len(raw), ErrIdentityKeyTooShort)
	}
	for _, other := range []string{"CMDCTRL_ADMIN_TOKEN", "CMDCTRL_SESSION_KEY"} {
		if v := strings.TrimSpace(getenv(other)); v != "" && v == raw {
			return nil, ErrIdentityKeyReused
		}
	}
	return NewSealer(raw)
}
