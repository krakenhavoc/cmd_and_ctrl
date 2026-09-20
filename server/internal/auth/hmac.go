package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SessionKeyEnv names the environment variable that holds the session
// signing key. It lives in /etc/cmd_and_ctrl/env beside
// CMDCTRL_ADMIN_TOKEN. See NewFromEnv for what happens when it is unset.
const SessionKeyEnv = "CMDCTRL_SESSION_KEY"

// MinSessionKeyBytes is the shortest signing key NewHMACAuthenticator
// accepts. 32 bytes is HMAC-SHA256's output size; a shorter key would
// be the weakest part of the construction.
const MinSessionKeyBytes = 32

// hmacTokenVersion prefixes every credential. It names the wire format,
// so a later change to the claims (ADR 0051 collapsing RolePlayer into
// RoleIdentified, say) can mint "v2." tokens and reject or migrate the
// old ones deliberately rather than misreading them.
const hmacTokenVersion = "v1"

// ErrSessionKeyTooShort is returned by NewHMACAuthenticator for a key
// under MinSessionKeyBytes.
var ErrSessionKeyTooShort = fmt.Errorf("auth: session key must be at least %d bytes", MinSessionKeyBytes)

// HMACAuthenticator is the stateless Authenticator (ADR 0044 decision
// 3). The Principal is serialised into the credential and signed with
// HMAC-SHA256, so Validate needs no server-side store. A token minted
// by one process validates in any later process started with the same
// key, which is the whole point: a deploy no longer logs everyone out.
//
// # Credential format
//
//	v1.<payload>.<signature>
//
// payload is the base64url (unpadded) encoding of a compact JSON claims
// object; signature is the base64url (unpadded) HMAC-SHA256 of the
// ASCII string "v1.<payload>". Every character is in the URL-safe
// unreserved set, so the credential goes into a WS `?token=`, a cookie
// or an Authorization header unescaped. A Discord-backed player session
// is roughly 400 characters, well inside the cookie and URL limits.
//
// The payload is signed, not encrypted. Anyone holding a token can read
// its claims (role, IDs, display name, Discord handle, timestamps); none
// of it is secret from the person the token was issued to. Anything
// that must stay server-side does not belong in a Principal.
//
// Timestamps are carried at millisecond precision, so the Principal
// Validate returns matches the one Issue returned exactly.
//
// # Revoke is advisory
//
// Revoke does nothing and returns nil. A stateless credential cannot be
// withdrawn by forgetting it: a token stays valid until its ExpiresAt,
// or until the signing key changes, which invalidates every session at
// once. ADR 0044 decision 3 accepts this cost. POST /logout still
// clears the cookie and the client drops its stored copy, so a logout
// ends the session in that browser; it does not kill a copy of the
// token held elsewhere. ADR 0051 decision 6 brings bounded revocation
// back per user, by comparing IssuedAt against
// users.sessions_invalid_before, which is why every token carries its
// issue time. That check is WithRevocation, which wraps this type; the
// HMAC authenticator itself stays stateless and knows nothing of it.
type HMACAuthenticator struct {
	key []byte
	now func() time.Time // injected so tests can control expiry
}

// NewHMACAuthenticator returns an authenticator signing with key. The
// key is copied. A key shorter than MinSessionKeyBytes is refused: this
// constructor never produces an authenticator with a weak or empty key.
func NewHMACAuthenticator(key []byte) (*HMACAuthenticator, error) {
	if len(key) < MinSessionKeyBytes {
		return nil, ErrSessionKeyTooShort
	}
	return &HMACAuthenticator{
		key: append([]byte(nil), key...),
		now: func() time.Time { return time.Now().UTC() },
	}, nil
}

// hmacClaims is the signed payload. Short keys keep the credential
// small, because it rides in a URL. uuid fields are strings so a zero
// ID is omitted rather than spelled out as 36 characters of zeros.
type hmacClaims struct {
	Role     Role   `json:"r"`
	UserID   string `json:"u,omitempty"`
	AdminID  string `json:"a,omitempty"`
	GameID   string `json:"g,omitempty"`
	PlayerID string `json:"p,omitempty"`
	Name     string `json:"n,omitempty"`

	DiscordID         string `json:"di,omitempty"`
	DiscordUsername   string `json:"du,omitempty"`
	DiscordGlobalName string `json:"dg,omitempty"`
	DiscordAvatarHash string `json:"da,omitempty"`

	IssuedAt  int64 `json:"iat"` // Unix milliseconds
	ExpiresAt int64 `json:"exp"` // Unix milliseconds
}

func uuidClaim(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

func parseUUIDClaim(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(s)
}

func claimsFor(p Principal) hmacClaims {
	return hmacClaims{
		Role:              p.Role,
		UserID:            uuidClaim(p.UserID),
		AdminID:           uuidClaim(p.AdminID),
		GameID:            uuidClaim(p.GameID),
		PlayerID:          uuidClaim(p.PlayerID),
		Name:              p.Name,
		DiscordID:         p.DiscordID,
		DiscordUsername:   p.DiscordUsername,
		DiscordGlobalName: p.DiscordGlobalName,
		DiscordAvatarHash: p.DiscordAvatarHash,
		IssuedAt:          p.IssuedAt.UnixMilli(),
		ExpiresAt:         p.ExpiresAt.UnixMilli(),
	}
}

func (c hmacClaims) principal() (Principal, error) {
	p := Principal{
		Role:              c.Role,
		Name:              c.Name,
		DiscordID:         c.DiscordID,
		DiscordUsername:   c.DiscordUsername,
		DiscordGlobalName: c.DiscordGlobalName,
		DiscordAvatarHash: c.DiscordAvatarHash,
		IssuedAt:          time.UnixMilli(c.IssuedAt).UTC(),
		ExpiresAt:         time.UnixMilli(c.ExpiresAt).UTC(),
	}
	var err error
	if p.UserID, err = parseUUIDClaim(c.UserID); err != nil {
		return Principal{}, err
	}
	if p.AdminID, err = parseUUIDClaim(c.AdminID); err != nil {
		return Principal{}, err
	}
	if p.GameID, err = parseUUIDClaim(c.GameID); err != nil {
		return Principal{}, err
	}
	if p.PlayerID, err = parseUUIDClaim(c.PlayerID); err != nil {
		return Principal{}, err
	}
	return p, nil
}

// sign returns the base64url HMAC-SHA256 of signingInput.
func (h *HMACAuthenticator) sign(signingInput string) []byte {
	mac := hmac.New(sha256.New, h.key)
	mac.Write([]byte(signingInput))
	return mac.Sum(nil)
}

// Issue stamps IssuedAt / ExpiresAt onto p, serialises it and signs it.
// A zero or negative ttl is a caller bug and is refused, as in
// MemoryAuthenticator.
func (h *HMACAuthenticator) Issue(_ context.Context, p Principal, ttl time.Duration) (string, Principal, error) {
	if ttl <= 0 {
		return "", Principal{}, ErrInvalidCredential
	}
	now := h.now().Truncate(time.Millisecond)
	p.IssuedAt = now
	p.ExpiresAt = now.Add(ttl).Truncate(time.Millisecond)

	payload, err := json.Marshal(claimsFor(p))
	if err != nil {
		return "", Principal{}, fmt.Errorf("auth: encode claims: %w", err)
	}
	signingInput := hmacTokenVersion + "." + base64.RawURLEncoding.EncodeToString(payload)
	sig := base64.RawURLEncoding.EncodeToString(h.sign(signingInput))
	return signingInput + "." + sig, p, nil
}

// Validate checks the signature in constant time before it reads a
// byte of the payload, then enforces ExpiresAt. Any malformed, foreign
// or tampered credential is ErrInvalidCredential; a genuine one past
// its expiry is ErrExpiredCredential.
func (h *HMACAuthenticator) Validate(_ context.Context, credential string) (Principal, error) {
	version, rest, ok := strings.Cut(credential, ".")
	if !ok || version != hmacTokenVersion {
		return Principal{}, ErrInvalidCredential
	}
	payloadB64, sigB64, ok := strings.Cut(rest, ".")
	if !ok || payloadB64 == "" || strings.Contains(sigB64, ".") {
		return Principal{}, ErrInvalidCredential
	}
	sig, err := base64.RawURLEncoding.Strict().DecodeString(sigB64)
	if err != nil {
		return Principal{}, ErrInvalidCredential
	}
	if !hmac.Equal(sig, h.sign(version+"."+payloadB64)) {
		return Principal{}, ErrInvalidCredential
	}

	payload, err := base64.RawURLEncoding.Strict().DecodeString(payloadB64)
	if err != nil {
		return Principal{}, ErrInvalidCredential
	}
	var c hmacClaims
	if err := json.Unmarshal(payload, &c); err != nil {
		return Principal{}, ErrInvalidCredential
	}
	if c.Role == "" || c.ExpiresAt == 0 {
		return Principal{}, ErrInvalidCredential
	}
	p, err := c.principal()
	if err != nil {
		return Principal{}, ErrInvalidCredential
	}
	if h.now().After(p.ExpiresAt) {
		return Principal{}, ErrExpiredCredential
	}
	return p, nil
}

// Revoke is advisory: it does nothing and returns nil. See the type's
// doc comment and ADR 0044 decision 3. A caller that needs a session
// dead must stop presenting it (POST /logout clears the cookie) or
// rotate the key.
func (h *HMACAuthenticator) Revoke(context.Context, string) error {
	return nil
}

var _ Authenticator = (*HMACAuthenticator)(nil)

// Logger is the one method NewFromEnv needs to warn. *slog.Logger
// satisfies it.
type Logger interface {
	Warn(msg string, args ...any)
}

// NewFromEnv picks the server's Authenticator from the environment
// variable SessionKeyEnv, read through getenv:
//
//   - Set to at least MinSessionKeyBytes bytes: an HMACAuthenticator.
//     Sessions survive a restart.
//   - Unset or blank: a MemoryAuthenticator, and a loud warning naming
//     the variable. A dev box keeps working; a production host without
//     the key visibly loses durability rather than quietly doing so.
//     This function never falls back to a fixed or derived key.
//   - Set but too short: an error. The operator asked for durable
//     sessions and gave a weak key; booting on the in-memory store
//     instead would hide that.
//
// Surrounding whitespace is trimmed, since env files are edited by
// hand. The key is otherwise used as raw bytes.
func NewFromEnv(getenv func(string) string, log Logger) (Authenticator, error) {
	raw := strings.TrimSpace(getenv(SessionKeyEnv))
	if raw == "" {
		log.Warn(SessionKeyEnv+" is not set; sessions are held IN MEMORY and every player is logged out on each restart or deploy. Set it to a random string of at least 32 bytes (e.g. `head -c 48 /dev/urandom | base64 -w0`) for durable sessions.",
			"var", SessionKeyEnv)
		return NewMemoryAuthenticator(), nil
	}
	a, err := NewHMACAuthenticator([]byte(raw))
	if err != nil {
		if errors.Is(err, ErrSessionKeyTooShort) {
			return nil, fmt.Errorf("%s is %d bytes: %w", SessionKeyEnv, len(raw), err)
		}
		return nil, err
	}
	return a, nil
}
