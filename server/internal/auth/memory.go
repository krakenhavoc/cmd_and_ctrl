package auth

import (
	"context"
	"sync"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/token"
)

// MemoryAuthenticator is the S04 default Authenticator: a simple
// server-side map from opaque credential → Principal. Tokens are 32
// random bytes, base64url-encoded.
//
// Properties at S04:
//   - Survives concurrent reads/writes via an internal mutex.
//   - Session is lost on server restart — users re-auth on reconnect.
//     Production runs HMACAuthenticator instead (S33, #517); this is
//     what NewFromEnv falls back to, loudly, when CMDCTRL_SESSION_KEY
//     is unset, and what the handler tests use.
//   - No rate limiting on Validate / Issue. A future HTTP middleware
//     can bolt this on if threat model warrants.
//   - Expired tokens are lazily dropped on Validate; there is no
//     background sweeper. For ≤a few dozen sessions this is fine.
//
// Revoke is real here: a revoked token fails Validate immediately. That
// is NOT true of HMACAuthenticator, so no caller may rely on it.
type MemoryAuthenticator struct {
	mu     sync.Mutex
	tokens map[string]Principal
	now    func() time.Time // injected so tests can control expiry
}

// NewMemoryAuthenticator constructs an empty MemoryAuthenticator.
func NewMemoryAuthenticator() *MemoryAuthenticator {
	return &MemoryAuthenticator{
		tokens: make(map[string]Principal),
		now:    func() time.Time { return time.Now().UTC() },
	}
}

// Issue generates a fresh random token, stores the principal keyed by
// that token, and returns the token along with the stored Principal
// (with IssuedAt / ExpiresAt filled in).
//
// ttl must be positive; a zero or negative TTL returns an error
// rather than minting an already-expired token — that's always a
// caller bug.
func (m *MemoryAuthenticator) Issue(ctx context.Context, p Principal, ttl time.Duration) (string, Principal, error) {
	if ttl <= 0 {
		return "", Principal{}, ErrInvalidCredential
	}
	tok, err := token.Random(32)
	if err != nil {
		return "", Principal{}, err
	}
	now := m.now()
	p.IssuedAt = now
	p.ExpiresAt = now.Add(ttl)

	m.mu.Lock()
	m.tokens[tok] = p
	m.mu.Unlock()
	return tok, p, nil
}

// Validate looks the token up in the in-memory store, rejects on
// expiry, and returns the Principal on success. An expired token is
// dropped from the store as a side effect — lazy cleanup keeps the
// map bounded without a sweeper goroutine.
func (m *MemoryAuthenticator) Validate(_ context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrInvalidCredential
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.tokens[token]
	if !ok {
		return Principal{}, ErrInvalidCredential
	}
	if m.now().After(p.ExpiresAt) {
		delete(m.tokens, token)
		return Principal{}, ErrExpiredCredential
	}
	return p, nil
}

// Revoke removes the token from the store. Returns nil whether the
// token existed or not — callers don't need to differentiate
// "already gone" from "just removed", and leaking that distinction
// would be a minor enumeration hazard.
func (m *MemoryAuthenticator) Revoke(_ context.Context, token string) error {
	m.mu.Lock()
	delete(m.tokens, token)
	m.mu.Unlock()
	return nil
}

// Count returns the number of currently-stored tokens. Used by tests
// and by the /healthz surface; never exposed on the wire to clients.
func (m *MemoryAuthenticator) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.tokens)
}

// compile-time assertion.
var _ Authenticator = (*MemoryAuthenticator)(nil)

// withContext is a no-op for the in-memory backend; future stateful
// backends (redis, sqlite) will use ctx for cancellation. Referenced
// here to document the contract.
var _ = context.Background
