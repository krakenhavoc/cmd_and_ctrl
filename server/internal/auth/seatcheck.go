package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SeatRemovedMessage is the 401 body message a request carries when its
// session's seat was removed from the table (#2785). The client matches
// on this exact string to tell a removed player from an expired
// session, so it is a constant and not prose.
const SeatRemovedMessage = "removed from table"

// SeatChecker answers one question for WithSeatCheck: has the seat a
// player session names been taken away from its holder?
//
// It is consulted on every Validate, so an implementation must answer
// from memory. lobby.Lobby is the production one.
type SeatChecker interface {
	// SeatRemoved reports whether a player session bound to
	// (gameID, playerID) is for a seat that is no longer at that
	// table. It must say false for a game it does not know, so a
	// table that is not loaded never turns into a sign-out.
	SeatRemoved(gameID, playerID uuid.UUID) bool
}

// WithSeatCheck wraps inner so that Validate also refuses a RolePlayer
// session whose seat has been removed from its table, with
// ErrSeatRemoved (#2785).
//
// Why this is its own wrapper and not Revoke: a session is signed, not
// stored (ADR 0044 decision 3), so Revoke is advisory and the token
// stays good until it expires. Per-user revocation (WithRevocation)
// would sign the person out of every table, and a guest or an agent
// seat has no user to revoke. The seat is the thing that went away, so
// the seat is what is checked. It needs no list of its own: the lobby's
// seat list is the truth, it survives a restart because the seats do,
// and a removed seat that is later re-created gets a new player id, so
// an old token can never match it. A seat removed from a STARTED game
// (#3093) stays in the list, so the engine's Player.Removed flag, carried
// by the snapshot, is what answers instead.
//
// Everything other than a player session passes straight through:
// admin, spectator and identified sessions name no seat.
func WithSeatCheck(inner Authenticator, seats SeatChecker) Authenticator {
	if seats == nil {
		return inner
	}
	return &seatCheckAuthenticator{inner: inner, seats: seats}
}

type seatCheckAuthenticator struct {
	inner Authenticator
	seats SeatChecker
}

func (a *seatCheckAuthenticator) Issue(ctx context.Context, p Principal, ttl time.Duration) (string, Principal, error) {
	return a.inner.Issue(ctx, p, ttl)
}

func (a *seatCheckAuthenticator) IssueUntil(ctx context.Context, p Principal, expiresAt time.Time) (string, Principal, error) {
	return a.inner.IssueUntil(ctx, p, expiresAt)
}

func (a *seatCheckAuthenticator) Validate(ctx context.Context, credential string) (Principal, error) {
	p, err := a.inner.Validate(ctx, credential)
	if err != nil {
		return Principal{}, err
	}
	if p.Role == RolePlayer && p.GameID != uuid.Nil && p.PlayerID != uuid.Nil &&
		a.seats.SeatRemoved(p.GameID, p.PlayerID) {
		return Principal{}, ErrSeatRemoved
	}
	return p, nil
}

func (a *seatCheckAuthenticator) Revoke(ctx context.Context, credential string) error {
	return a.inner.Revoke(ctx, credential)
}

var _ Authenticator = (*seatCheckAuthenticator)(nil)
