// Package auth defines the identity / authorization boundary for
// cmd_and_ctrl. The design intent is explicit pluggability: lobby
// handlers and the WebSocket hub speak only in terms of Principal
// and the Authenticator interface, never raw tokens or cookies.
// Different authenticator implementations can plug into the same
// seam: MemoryAuthenticator, the S04 in-memory token store, and
// HMACAuthenticator, the stateless signed credential that survives a
// restart (ADR 0044 decision 3, S33). NewFromEnv picks between them.
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
//   - RoleIdentified: a Discord-authenticated visitor with no seat
//     and no game. Minted by the login-page OAuth flow — the one
//     started without an invite in hand — and good for exactly one
//     thing: POST /join, which trades it plus an invite code for a
//     RolePlayer session. It deliberately cannot open a WebSocket;
//     lobby.WSAuthorizer rejects it by name rather than letting it
//     fall through, because a principal arriving at the hub with a
//     nil GameID would otherwise bind to the zero game. Added
//     Sept 2026.
type Role string

const (
	RolePlayer     Role = "player"
	RoleAdmin      Role = "admin"
	RoleSpectator  Role = "spectator"
	RoleIdentified Role = "identified"
)

// Principal is the canonical authenticated identity. It is the only
// thing the rest of the server cares about — the shape of the
// underlying credential (an opaque random token, or a signed payload
// from HMACAuthenticator) is never leaked past the Authenticator
// boundary. HMACAuthenticator serialises every field below into the
// token, so a new field here must be added to its claims too
// (hmac.go), or it silently comes back zero after a Validate.
//
// A RolePlayer principal MUST have a non-zero PlayerID and GameID. A
// RoleAdmin principal has an AdminID for audit-log identification
// but typically no GameID until it claims a seat. A RoleIdentified
// principal carries the Discord* fields and neither GameID nor
// PlayerID — it is an identity waiting for an invite code.
type Principal struct {
	Role Role `json:"role"`
	// UserID is the users-table row this session belongs to (ADR 0051
	// decision 3). Set on the RoleIdentified session the Discord
	// callback mints, and carried onto every RolePlayer session minted
	// from it (the callback's invite flow, and POST /join). Zero for
	// admin, spectator and guest sessions, and zero for everyone on a
	// deployment with no database (CMDCTRL_DATA_DIR empty).
	//
	// UserID is also what makes a session revocable (ADR 0051 decision
	// 6): WithRevocation refuses a principal whose IssuedAt is at or
	// before its user's sessions_invalid_before. A principal with a
	// zero UserID — admin, spectator, guest, or any session on a
	// deployment with no database — is never checked. It keeps exactly
	// the ADR 0044 decision 3 posture: a short TTL and an advisory
	// Revoke. There is no way to withdraw one early short of rotating
	// CMDCTRL_SESSION_KEY.
	UserID   uuid.UUID `json:"user_id,omitempty"`
	AdminID  uuid.UUID `json:"admin_id,omitempty"`
	GameID   uuid.UUID `json:"game_id,omitempty"`
	PlayerID uuid.UUID `json:"player_id,omitempty"`
	Name     string    `json:"name,omitempty"`

	// Discord* fields are populated when a RolePlayer session is
	// minted via the OAuth flow (S12.5). All four are optional —
	// a player who joins with manual name entry has zero values
	// across the board. DiscordID is the stable snowflake; the
	// other three are cached at sign-in time so the server can
	// render the seat label + avatar without re-hitting the
	// Discord API on every snapshot. Avatar hash is cached here
	// rather than joined from a separate store because it's
	// small and changes rarely.
	DiscordID         string `json:"discord_id,omitempty"`
	DiscordUsername   string `json:"discord_username,omitempty"`
	DiscordGlobalName string `json:"discord_global_name,omitempty"`
	DiscordAvatarHash string `json:"discord_avatar_hash,omitempty"`

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
	// ErrRevokedCredential is a genuine, unexpired credential whose
	// user has since signed out everywhere or been removed by an
	// admin (ADR 0051 decision 6). Only WithRevocation returns it.
	ErrRevokedCredential = errors.New("auth: revoked credential")
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
// Stateful implementations (MemoryAuthenticator) keep a map of
// credential→Principal internally. Stateless implementations
// (HMACAuthenticator) serialise the Principal into the credential
// itself and verify the signature on Validate.
type Authenticator interface {
	Issue(ctx context.Context, p Principal, ttl time.Duration) (credential string, issued Principal, err error)
	Validate(ctx context.Context, credential string) (Principal, error)

	// Revoke optionally invalidates a specific credential ahead of its
	// expiry. Stateless backends may return nil without doing
	// anything; stateful backends should drop the token from their
	// store. Callers that need guaranteed revocation should check the
	// documentation of the specific Authenticator they're using:
	// MemoryAuthenticator's Revoke is real, HMACAuthenticator's is
	// advisory (a no-op, ADR 0044 decision 3).
	Revoke(ctx context.Context, credential string) error
}
