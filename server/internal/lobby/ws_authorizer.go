package lobby

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// WSAuthorizer is the glue between the auth package and the hub's
// UpgradeAuthorizer seam. It:
//
//  1. Pulls the credential off the request (cookie > bearer > query).
//  2. Validates it with the configured Authenticator.
//  3. Cross-checks the principal against the requested ?game= (if any)
//     and returns the (gameID, playerID) pair the hub should bind to.
//
// A RolePlayer principal is locked to exactly the (gameID, playerID)
// pair minted at join time — it cannot spy on a different game. An
// admin is permitted to bind to any game, optionally as any seat
// (?player=); omitting it yields the admin's omniscient spectator view.
//
// "An admin" is isAdminPrincipal: the shared-token session, or a
// signed-in person on CMDCTRL_DISCORD_ADMIN_USER_IDS (ADR 0110 §3,
// owner answer 4: full parity with the token). For a signed-in admin:
//
//   - A player or spectator session at its own game, asking for no
//     other seat, binds exactly as it would for anyone, with
//     Binding.Admin set so the table-host gates know.
//   - Any other request — a different ?game=, a ?player= that is not
//     the session's own seat, or an identified session with no seat at
//     all — takes the admin branch: any game, optionally as any seat.
//   - A non-admin is unchanged.
//
// Every admin binding is logged at Info with who it is (admin_user_id,
// or admin_id for the token), the game and the seat bound, if any.
//
// Every binding also carries the session's UserID and IssuedAt, so a
// logout-everywhere or an admin revoke can close the sockets that
// session opened (ws.Hub.EvictUserSessions, ADR 0051 decision 6). The
// revocation check itself is not here: it is in Auth, which main wraps
// with auth.WithRevocation, so a revoked token fails Validate below
// like an expired one.
type WSAuthorizer struct {
	Auth auth.Authenticator
	// Admins is the same allowlist lobby.Config.Admins holds. Nil is
	// the empty list: only the shared token is an admin.
	Admins *AdminList
	// Log receives the admin-binding audit lines. Nil: none.
	Log *slog.Logger
}

// AuthorizeUpgrade implements ws.UpgradeAuthorizer.
func (a *WSAuthorizer) AuthorizeUpgrade(r *http.Request) (ws.Binding, error) {
	cred := auth.CredentialFromRequest(r)
	if cred == "" {
		return ws.Binding{}, ws.StatusError(http.StatusUnauthorized, "authentication required")
	}
	p, err := a.Auth.Validate(r.Context(), cred)
	if err != nil {
		return ws.Binding{}, ws.StatusError(http.StatusUnauthorized, err.Error())
	}

	q := r.URL.Query()
	var requestedGame uuid.UUID
	if raw := q.Get("game"); raw != "" {
		id, perr := uuid.Parse(raw)
		if perr != nil {
			return ws.Binding{}, ws.StatusError(http.StatusBadRequest, "invalid game id")
		}
		requestedGame = id
	}
	var requestedPlayer uuid.UUID
	if raw := q.Get("player"); raw != "" {
		id, perr := uuid.Parse(raw)
		if perr != nil {
			return ws.Binding{}, ws.StatusError(http.StatusBadRequest, "invalid player id")
		}
		requestedPlayer = id
	}

	if isAdminPrincipal(a.Admins, p) {
		if !isServerCredential(p) && a.ownBinding(p, requestedGame, requestedPlayer) {
			b, err := bindingFor(p, requestedGame)
			if err != nil {
				return ws.Binding{}, err
			}
			b.Admin = true
			a.logAdminBinding(p, b, "own_session")
			return b, nil
		}
		return a.adminBinding(p, requestedGame, requestedPlayer)
	}
	return bindingFor(p, requestedGame)
}

// ownBinding reports whether a signed-in admin's request is just their
// own session's ordinary binding: a seat or spectator session at its
// own game (or no ?game= at all), naming no other seat.
func (a *WSAuthorizer) ownBinding(p auth.Principal, game, player uuid.UUID) bool {
	if p.GameID == uuid.Nil {
		return false
	}
	if p.Role != auth.RolePlayer && p.Role != auth.RoleSpectator {
		return false
	}
	if game != uuid.Nil && game != p.GameID {
		return false
	}
	return player == uuid.Nil || player == p.PlayerID
}

// adminBinding is the moderator escape hatch: any game, as the named
// seat or (uuid.Nil) the omniscient spectator view. Admins are NOT
// marked read-only — they need to drive state on a player's behalf.
func (a *WSAuthorizer) adminBinding(p auth.Principal, game, player uuid.UUID) (ws.Binding, error) {
	if game == uuid.Nil {
		return ws.Binding{}, ws.StatusError(http.StatusBadRequest, "missing game id")
	}
	b := ws.Binding{GameID: game, PlayerID: player, UserID: p.UserID, IssuedAt: p.IssuedAt, Admin: true}
	a.logAdminBinding(p, b, "any_game")
	return b, nil
}

func (a *WSAuthorizer) logAdminBinding(p auth.Principal, b ws.Binding, branch string) {
	if a.Log == nil {
		return
	}
	seat := ""
	if b.PlayerID != uuid.Nil {
		seat = b.PlayerID.String()
	}
	a.Log.Info("admin websocket binding",
		append(adminWho(p), "game_id", b.GameID.String(), "seat", seat, "read_only", b.ReadOnly, "branch", branch)...)
}

// bindingFor is the non-admin binding for p's own session.
func bindingFor(p auth.Principal, requestedGame uuid.UUID) (ws.Binding, error) {
	switch p.Role {
	case auth.RolePlayer:
		// Player sessions are minted bound to one game. The query
		// string may omit ?game= (server resolves it from the
		// principal) or, if present, MUST match — a mismatch is a
		// sign of a copy-paste invite leak or a client bug.
		if requestedGame != uuid.Nil && requestedGame != p.GameID {
			return ws.Binding{}, ws.StatusError(http.StatusForbidden, "session is not for this game")
		}
		return ws.Binding{GameID: p.GameID, PlayerID: p.PlayerID, UserID: p.UserID, IssuedAt: p.IssuedAt}, nil

	case auth.RoleSpectator:
		// Spectator sessions are minted bound to one game (no player).
		// Same query-string handling as RolePlayer (?game= must match
		// or be omitted), but the bound playerID is always uuid.Nil
		// — the hub then treats them like a spectator for visibility
		// filtering, while ReadOnly = true makes the action-frame gate
		// refuse any mutation frame they send.
		if requestedGame != uuid.Nil && requestedGame != p.GameID {
			return ws.Binding{}, ws.StatusError(http.StatusForbidden, "session is not for this game")
		}
		return ws.Binding{GameID: p.GameID, ReadOnly: true, UserID: p.UserID, IssuedAt: p.IssuedAt}, nil

	case auth.RoleIdentified:
		// A Discord sign-in that hasn't claimed a seat yet: no game,
		// no player, nothing to bind to. Refused by name rather than
		// through the default arm for two reasons — the message can
		// tell the client what to do next (POST /join with an invite
		// code, which swaps this for a RolePlayer session), and the
		// default arm's "unrecognised role" would be a lie about a
		// role the server mints itself.
		return ws.Binding{}, ws.StatusError(http.StatusForbidden,
			"sign-in session has no seat yet — join a table with an invite code first")

	default:
		return ws.Binding{}, ws.StatusError(http.StatusForbidden, "unrecognised principal role")
	}
}

// compile-time assertion.
var _ ws.UpgradeAuthorizer = (*WSAuthorizer)(nil)
