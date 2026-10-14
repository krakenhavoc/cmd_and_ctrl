package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RevocationList answers ADR 0051 decision 6's one question: has this
// user withdrawn every session issued at or before a given time?
//
// It is consulted on every Validate, which means every authenticated
// HTTP request and every WebSocket upgrade, so an implementation must
// answer from memory. users.Revocations is the production one: the
// users.sessions_invalid_before column, loaded once at boot and
// updated in place by its own writes. This package deliberately knows
// nothing about where the list is kept; it has no database import and
// should not grow one.
type RevocationList interface {
	// Revoked reports whether a session for userID issued at issuedAt
	// has been withdrawn. userID is never uuid.Nil here.
	Revoked(userID uuid.UUID, issuedAt time.Time) bool
}

// WithRevocation wraps inner so that Validate also refuses any
// principal with a non-zero UserID whose user has revoked it, with
// ErrRevokedCredential. Everything else passes straight through:
//
//   - Issue is inner's, except that a token minted in the same
//     millisecond as its user's watermark is minted again once the clock
//     has passed it, so a fresh sign-in is never born revoked. Whether a
//     token is still good is otherwise decided when it is presented.
//   - Revoke is inner's, so still advisory for HMAC sessions. Per-user
//     revocation is a different operation (users.Revocations.RevokeAll)
//     because it withdraws every session a person holds, not one token.
//   - A principal with a zero UserID (admin, spectator, guest) is
//     returned unchecked. See Principal.UserID.
//
// Expiry and signature are checked first, by inner, so a revoked
// credential that has also expired reports ErrExpiredCredential.
//
// A nil list returns inner unchanged, which is how a deployment with no
// database runs.
func WithRevocation(inner Authenticator, list RevocationList) Authenticator {
	if list == nil {
		return inner
	}
	return &revokingAuthenticator{inner: inner, list: list, pause: func() { time.Sleep(time.Millisecond) }}
}

type revokingAuthenticator struct {
	inner Authenticator
	list  RevocationList
	// pause waits out one issue-time tick (a millisecond). Injected so
	// tests with a fake clock can move it on.
	pause func()
}

// maxTieRetries bounds how long a mint waits for the clock to pass its
// user's watermark. One retry is enough for a real clock; the cap only
// stops a clock that is behind the watermark from looping.
const maxTieRetries = 5

// mint issues a token with mk and makes sure it is not born revoked.
//
// IssuedAt and the watermark are both whole milliseconds, and Revoked
// refuses "at or before" the watermark, so a token minted in the very
// millisecond of a revocation is refused. It must be: a token minted
// just before the revoke in that millisecond is indistinguishable from
// it, so the comparison cannot be loosened. Instead a session minted
// into a tie is thrown away and minted again once the clock has moved
// on, which makes its IssuedAt strictly after the watermark. Every token
// that validates was therefore issued after the revoke, and every token
// issued before it is still refused. The discarded token was never
// returned, so nothing is leaked.
func (a *revokingAuthenticator) mint(p Principal, mk func() (string, Principal, error)) (string, Principal, error) {
	for i := 0; ; i++ {
		tok, issued, err := mk()
		if err != nil || p.UserID == uuid.Nil || i >= maxTieRetries ||
			!a.list.Revoked(p.UserID, issued.IssuedAt) {
			return tok, issued, err
		}
		a.pause()
	}
}

func (a *revokingAuthenticator) Issue(ctx context.Context, p Principal, ttl time.Duration) (string, Principal, error) {
	return a.mint(p, func() (string, Principal, error) { return a.inner.Issue(ctx, p, ttl) })
}

func (a *revokingAuthenticator) IssueUntil(ctx context.Context, p Principal, expiresAt time.Time) (string, Principal, error) {
	return a.mint(p, func() (string, Principal, error) { return a.inner.IssueUntil(ctx, p, expiresAt) })
}

func (a *revokingAuthenticator) Validate(ctx context.Context, credential string) (Principal, error) {
	p, err := a.inner.Validate(ctx, credential)
	if err != nil {
		return Principal{}, err
	}
	if p.UserID != uuid.Nil && a.list.Revoked(p.UserID, p.IssuedAt) {
		return Principal{}, ErrRevokedCredential
	}
	return p, nil
}

func (a *revokingAuthenticator) Revoke(ctx context.Context, credential string) error {
	return a.inner.Revoke(ctx, credential)
}

var _ Authenticator = (*revokingAuthenticator)(nil)
