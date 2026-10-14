package lobby

// stack_position_http.go is the HTTP half of the movable stack panel
// (stack_position.go, #3429, ADR 0176):
//
//	PUT    /games/{id}/stack-position   {"x": 0..1, "y": 0..1}: save the spot
//	DELETE /games/{id}/stack-position                          : back to the default
//
// The caller is a seat at the table, by its session, and the seat is the
// session's, never one named in the request, so nobody can place another
// seat's stack. A spectator or the admin token is refused. Both answer 204;
// the new spot reaches the seat's clients on the next state frame
// (GameView.stack_positions).

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// stackPositionCaller is the seat a stack-position route is called from: a
// player session bound to this game, never the admin token or a spectator.
func stackPositionCaller(r *http.Request) (uuid.UUID, auth.Principal, error) {
	id, err := gameIDFromPath(r)
	if err != nil {
		return uuid.Nil, auth.Principal{}, err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return uuid.Nil, auth.Principal{}, httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.Role != auth.RolePlayer || p.GameID != id || p.PlayerID == uuid.Nil {
		return uuid.Nil, auth.Principal{}, httpError(http.StatusForbidden, "only a seat at this table can place its stack")
	}
	return id, p, nil
}

// putStackPosition handles PUT /games/{id}/stack-position.
func putStackPosition(c Config, w http.ResponseWriter, r *http.Request) error {
	id, p, err := stackPositionCaller(r)
	if err != nil {
		return err
	}
	var req struct{ X, Y *float64 }
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return httpError(http.StatusBadRequest, "invalid body: "+err.Error())
	}
	if req.X == nil || req.Y == nil {
		return httpError(http.StatusBadRequest, "x and y are both required")
	}
	if err := c.Lobby.SetStackPosition(id, p.PlayerID, &protocol.StackPositionView{X: *req.X, Y: *req.Y}); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// deleteStackPosition handles DELETE /games/{id}/stack-position.
func deleteStackPosition(c Config, w http.ResponseWriter, r *http.Request) error {
	id, p, err := stackPositionCaller(r)
	if err != nil {
		return err
	}
	if err := c.Lobby.SetStackPosition(id, p.PlayerID, nil); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
