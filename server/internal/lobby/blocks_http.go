package lobby

// blocks_http.go: the block list and the unblock (#3113, ADR 0075
// section 2.6 amendment).
//
//	GET    /games/{id}/blocks              who this table has blocked
//	DELETE /games/{id}/blocks/{block_id}   lift one block
//
// WHO: the same authority split as removing a seat (seat_remove.go).
// The table host, the game's creator and an admin on a table that has
// not started; an admin alone on a started or ended one. The host and
// the creator get a 409 there, as they do on DELETE /games/{id}/seats/
// {player}, and a seat that is neither gets 403. The admin sees and
// lifts a block on any table from the Games view or the lobby.
//
// WHAT IT SHOWS: one line per removal, labelled by the account's display
// name or by the guest's folded name, and whether the block holds
// against an account or only a name. Never a Discord id, never a user id.
// A removed spectator who was not also blocked is not a line: there is
// nothing to lift.

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blockListResponse is both routes' answer: the table's blocks as they
// now stand.
type blockListResponse struct {
	Blocks []BlockEntry `json:"blocks"`
}

// authorizeBlockManager is the gate both routes share: the caller must
// be able to manage seats here, and a table that has begun is the
// admin's alone.
func authorizeBlockManager(c Config, r *http.Request) (uuid.UUID, error) {
	id, err := gameIDFromPath(r)
	if err != nil {
		return uuid.Nil, err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return uuid.Nil, httpError(http.StatusInternalServerError, "missing principal")
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return uuid.Nil, err
	}
	admin := c.isAdmin(p)
	if !CanRemoveSeat(p, meta, admin) {
		return uuid.Nil, ErrNotSeatManager
	}
	if meta.State != string(game.StateLobby) && !admin {
		return uuid.Nil, ErrGameStarted
	}
	return id, nil
}

// listBlocks handles GET /games/{id}/blocks.
//
//	401  anonymous
//	403  not the host, the creator or an admin
//	404  no such table
//	409  the table has started and the caller is not an admin
//	200  {"blocks": [{"id", "label", "kind": "account"|"guest", "created_at"}]}
func listBlocks(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := authorizeBlockManager(c, r)
	if err != nil {
		return err
	}
	list, err := c.Lobby.BlockList(id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, blockListResponse{Blocks: nonNilBlocks(list)})
}

// unblock handles DELETE /games/{id}/blocks/{block_id}.
//
//	401  anonymous
//	403  not the host, the creator or an admin
//	404  no such table, or no such block on it
//	409  the table has started and the caller is not an admin
//	200  {"blocks": [...]}  the list without it
func unblock(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := authorizeBlockManager(c, r)
	if err != nil {
		return err
	}
	blockID := r.PathValue("block_id")
	if blockID == "" {
		return httpError(http.StatusBadRequest, "missing block id")
	}
	if err := c.Lobby.Unblock(id, blockID); err != nil {
		if errors.Is(err, ErrBlockNotFound) {
			return httpError(http.StatusNotFound, "no such block at this table")
		}
		return err
	}
	c.logger().Info("table block lifted", "game_id", id, "block_id", blockID)
	list, err := c.Lobby.BlockList(id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, blockListResponse{Blocks: nonNilBlocks(list)})
}

func nonNilBlocks(b []BlockEntry) []BlockEntry {
	if b == nil {
		return []BlockEntry{}
	}
	return b
}
