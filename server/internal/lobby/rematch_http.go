package lobby

// rematch_http.go is the HTTP half of the rematch (rematch.go, #3285, ADR
// 0164):
//
//	POST /games/{id}/rematch         — the ended table's host: offer one
//	POST /games/{id}/rematch/accept  — another human seat: take a place
//	POST /games/{id}/rematch/leave   — another human seat: decline
//
// The caller is a seat at the OLD table, by its session. Offering and
// accepting answer with a session for the caller's seat at the NEW table,
// like a join does, and set the session cookie to it: the client swaps to
// it and goes to the new table. The new table is started, filled with
// bots and so on through the routes every table has.

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// rematchCaller is the seat a rematch route is called from: a player
// session bound to this game, never the admin token.
func rematchCaller(r *http.Request) (uuid.UUID, auth.Principal, error) {
	id, err := gameIDFromPath(r)
	if err != nil {
		return uuid.Nil, auth.Principal{}, err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return uuid.Nil, auth.Principal{}, httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.Role != auth.RolePlayer || p.GameID != id || p.PlayerID == uuid.Nil {
		return uuid.Nil, auth.Principal{}, httpError(http.StatusForbidden, "only a seat at this table can answer its rematch")
	}
	return id, p, nil
}

// offerRematch handles POST /games/{id}/rematch.
func offerRematch(c Config, w http.ResponseWriter, r *http.Request) error {
	id, p, err := rematchCaller(r)
	if err != nil {
		return err
	}
	res, err := c.Lobby.OfferRematch(id, p.PlayerID)
	if err != nil {
		return err
	}
	metrics.GameCreated()
	metrics.RematchOffered()
	c.logger().Info("rematch offered", "game_id", id.String(), "rematch_id", res.Game.ID.String())
	return writeRematchSession(c, w, r, p, res, http.StatusCreated)
}

// acceptRematch handles POST /games/{id}/rematch/accept.
func acceptRematch(c Config, w http.ResponseWriter, r *http.Request) error {
	id, p, err := rematchCaller(r)
	if err != nil {
		return err
	}
	res, err := c.Lobby.AcceptRematch(id, p.PlayerID)
	if err != nil {
		return err
	}
	return writeRematchSession(c, w, r, p, res, http.StatusOK)
}

// leaveRematch handles POST /games/{id}/rematch/leave.
func leaveRematch(c Config, w http.ResponseWriter, r *http.Request) error {
	id, p, err := rematchCaller(r)
	if err != nil {
		return err
	}
	if err := c.Lobby.DeclineRematch(id, p.PlayerID); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// writeRematchSession mints the caller's seat session for the new table,
// as a join does: bound to the new game and seat, carrying the caller's
// user and Discord identity, and expiring with the session it came from
// for a signed-in person (ADR 0110 §1). Neither invite is returned: the
// table has none.
func writeRematchSession(c Config, w http.ResponseWriter, r *http.Request, old auth.Principal, res RematchResult, status int) error {
	seat := withIdentity(auth.Principal{
		Role:     auth.RolePlayer,
		UserID:   old.UserID,
		GameID:   res.Game.ID,
		PlayerID: res.PlayerID,
		Name:     old.Name,
	}, principalIdentity(old))
	var source *auth.Principal
	if old.UserID != uuid.Nil {
		source = &old
	}
	tok, issued, err := issueFor(r.Context(), c, seat, source)
	if err != nil {
		return err
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)
	meta := res.Game
	meta.InviteToken = ""
	meta.SpectatorInvite = ""
	return writeJSON(w, status, sessionResponse{
		Token:     tok,
		ExpiresAt: issued.ExpiresAt,
		Principal: issued,
		Game:      &meta,
		PlayerID:  res.PlayerID,
	})
}
