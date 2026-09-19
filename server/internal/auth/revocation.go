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
//   - Issue is inner's. A token is always minted; whether it is still
//     good is decided when it is presented.
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
	return &revokingAuthenticator{inner: inner, list: list}
}

type revokingAuthenticator struct {
	inner Authenticator
	list  RevocationList
}

func (a *revokingAuthenticator) Issue(ctx context.Context, p Principal, ttl time.Duration) (string, Principal, error) {
	return a.inner.Issue(ctx, p, ttl)
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
