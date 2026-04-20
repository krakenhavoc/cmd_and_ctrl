// Package auth defines the identity / authorization boundary for
// cmd_and_ctrl. The design intent is explicit pluggability: lobby
// handlers and the WebSocket hub speak only in terms of Principal
// and the Authenticator interface, never raw tokens or cookies.
// Different authenticator implementations can plug into the same
// seam — the S04 default is a stateful in-memory invite-token store,
// and the post-S04 roadmap includes a stateless HMAC implementation
// that needs zero server-side bookkeeping.
//
// Transport decisions live outside this package: the HTTP middleware
// and WebSocket authorizer (see http.go and the ws.UpgradeAuthorizer
// hook) call Issue / Validate and handle cookie / query-param wiring
// themselves.
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role tags what a Principal is allowed to do.
//
//   - RolePlayer: a seated player in a specific game. Allowed to
//     open WS connections to that game only, and to hit /games/:id
//     routes for that game.
//   - RoleAdmin: whoever runs the server. Allowed to create games,
//     list all games, and claim any seat. There is no admin-only
//     "mutate game state" endpoint — admins act as players once
//     they've claimed a seat.
//   - RoleSpectator: a non-seated observer of a specific game.
//     Receives snapshot broadcasts (filtered — all opponent hands
//     hidden, same as admin spectator) but cannot send action
//     frames. Bound to GameID only; PlayerID is uuid.Nil. Issued
//     via the per-game spectator invite from the lobby. Added in
//     S11.
type Role string

const (
	RolePlayer    Role = "player"
	RoleAdmin     Role = "admin"
	RoleSpectator Role = "spectator"
)

// Principal is the canonical authenticated identity. It is the only
// thing the rest of the server cares about — the shape of the
// underlying credential (opaque token today, JWT tomorrow) is never
// leaked past the Authenticator boundary.
//
// A RolePlayer principal MUST have a non-zero PlayerID and GameID. A
// RoleAdmin principal has an AdminID for audit-log identification
// but typically no GameID until it claims a seat.
type Principal struct {
	Role     Role      `json:"role"`
	AdminID  uuid.UUID `json:"admin_id,omitempty"`
	GameID   uuid.UUID `json:"game_id,omitempty"`
	PlayerID uuid.UUID `json:"player_id,omitempty"`
	Name     string    `json:"name,omitempty"`

	// IssuedAt / ExpiresAt are populated by Issue and checked by
	// Validate. They're exposed so clients (and tests) can tell when
	// a session will expire without holding the Authenticator.
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Common errors surfaced by Authenticator implementations. These are
// sentinel errors so callers (the HTTP middleware and the WS
// authorizer) can map them to specific HTTP statuses without caring
// which Authenticator produced them.
var (
	ErrInvalidCredential = errors.New("auth: invalid credential")
	ErrExpiredCredential = errors.New("auth: expired credential")
	ErrUnknownPrincipal  = errors.New("auth: unknown principal")
)

// Authenticator mints and validates credentials. Implementations must
// be safe for concurrent use.
//
// The two operations are intentionally dual:
//
//   - Issue: given a Principal (newly claimed seat, admin login), mint
//     a credential string the caller can ship to the client. The
//     returned Principal has IssuedAt / ExpiresAt populated so the
//     caller can tell the client when the session will expire.
//   - Validate: given a credential string, return the Principal it
//     authenticates or one of the sentinel errors above.
//
// Stateful implementations (the S04 invite store) keep a map of
// credential→Principal internally. Stateless implementations
// (HMAC-signed tokens, future) serialise the Principal into the
// credential itself and verify the signature on Validate.
type Authenticator interface {
	Issue(ctx context.Context, p Principal, ttl time.Duration) (credential string, issued Principal, err error)
	Validate(ctx context.Context, credential string) (Principal, error)

	// Revoke optionally invalidates a specific credential ahead of its
	// expiry. Stateless backends (HMAC) may return nil without doing
	// anything; stateful backends should drop the token from their
	// store. Callers that need guaranteed revocation should check the
	// documentation of the specific Authenticator they're using.
	Revoke(ctx context.Context, credential string) error
}
