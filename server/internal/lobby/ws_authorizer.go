package lobby

import (
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
// pair minted at join time — it cannot spy on a different game. A
// RoleAdmin principal is permitted to bind to any game but must
// supply ?player= explicitly; omitting it yields a spectator view.
type WSAuthorizer struct {
	Auth auth.Authenticator
}

// AuthorizeUpgrade implements ws.UpgradeAuthorizer.
func (a *WSAuthorizer) AuthorizeUpgrade(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	cred := auth.CredentialFromRequest(r)
	if cred == "" {
		return uuid.Nil, uuid.Nil, ws.StatusError(http.StatusUnauthorized, "authentication required")
	}
	p, err := a.Auth.Validate(r.Context(), cred)
	if err != nil {
		return uuid.Nil, uuid.Nil, ws.StatusError(http.StatusUnauthorized, err.Error())
	}

	q := r.URL.Query()
	var requestedGame uuid.UUID
	if raw := q.Get("game"); raw != "" {
		id, perr := uuid.Parse(raw)
		if perr != nil {
			return uuid.Nil, uuid.Nil, ws.StatusError(http.StatusBadRequest, "invalid game id")
		}
		requestedGame = id
	}

	switch p.Role {
	case auth.RolePlayer:
		// Player sessions are minted bound to one game. The query
		// string may omit ?game= (server resolves it from the
		// principal) or, if present, MUST match — a mismatch is a
		// sign of a copy-paste invite leak or a client bug.
		if requestedGame != uuid.Nil && requestedGame != p.GameID {
			return uuid.Nil, uuid.Nil, ws.StatusError(http.StatusForbidden, "session is not for this game")
		}
		return p.GameID, p.PlayerID, nil

	case auth.RoleAdmin:
		// Admins bind to whatever game they asked for. ?player= is
		// optional; uuid.Nil falls back to spectator view.
		if requestedGame == uuid.Nil {
			return uuid.Nil, uuid.Nil, ws.StatusError(http.StatusBadRequest, "missing game id")
		}
		var playerID uuid.UUID
		if raw := q.Get("player"); raw != "" {
			id, perr := uuid.Parse(raw)
			if perr != nil {
				return uuid.Nil, uuid.Nil, ws.StatusError(http.StatusBadRequest, "invalid player id")
			}
			playerID = id
		}
		return requestedGame, playerID, nil

	default:
		return uuid.Nil, uuid.Nil, ws.StatusError(http.StatusForbidden, "unrecognised principal role")
	}
}

// compile-time assertion.
var _ ws.UpgradeAuthorizer = (*WSAuthorizer)(nil)
