package lobby

// seat_remove.go: the host removes someone else's seat from an
// unstarted table (#2785).
//
// Until this existed only a bot seat could be taken off a table
// (DELETE /games/{id}/seats/bot/{player}). A human or agent guest who
// sat down under the wrong name could not be moved, and the only way
// out was a new table and a new invite.
//
// SCOPE: unstarted tables only. Removing a seat from a game that has
// started needs a CR-sound answer for that seat's permanents and
// objects (concede them out, CR 800.4a), which is the owner's call and
// is not built here. The route refuses a started table with 409, and
// the engine's RemovePlayer refuses it again underneath.
//
// WHO: the table host, the game's creator, or an admin. That is the
// set that manages the table (CanManageTable) and the invites
// (CanRotateInvites) put together. It is narrower than the bot route,
// which any seated player may call: removing a person is not the same
// kind of act as adding a bot, and no one gains it who did not have it.
//
// WHAT A REMOVAL DOES:
//
//   - the seat leaves the engine, so seat numbers close up, the slot is
//     free for the next join, and every connected client gets the new
//     lobby state;
//   - the seats row is rewritten;
//   - the seat's session stops validating (auth.WithSeatCheck asks
//     Lobby.SeatRemoved), so a reclaim, a reconnect or any other call
//     with that credential is refused;
//   - the HTTP layer closes the seat's sockets with close reason
//     ws.SeatRemovedReason, which is how the removed client learns why.
//
// The invite is not touched. A removed person who still holds the
// invite link can sit down again, as a new seat. Rotating the invite is
// the way to stop that, and the Remove control says so.

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var (
	// ErrNotSeatManager is returned by DELETE /games/{id}/seats/{player}
	// when the caller is not the host, the creator or an admin.
	ErrNotSeatManager = errors.New("lobby: only the table host, its creator or the admin may remove a seat")

	// ErrCannotRemoveHost refuses removing the host's own seat. Hand the
	// table to someone else first (POST /games/{id}/host).
	ErrCannotRemoveHost = errors.New("lobby: the host's seat cannot be removed; transfer the host first")

	// ErrSeatIsBotRoute refuses a bot seat on the human-seat route: bots
	// have their own, with its own, wider, authorisation.
	ErrSeatIsBotRoute = errors.New("lobby: that seat is a bot; remove it with DELETE /games/{id}/seats/bot/{player}")
)

// SeatEvictor closes the WebSockets of one removed seat. *ws.Hub
// implements it.
type SeatEvictor interface {
	EvictSeat(gameID, playerID uuid.UUID) int
}

// CanRemoveSeat reports whether p may remove a seat from the table meta
// describes: an admin, the seated host, or the signed-in creator.
// meta must be fresh from Lobby.Get (which resolves the effective host).
func CanRemoveSeat(p auth.Principal, meta GameMeta, admin bool) bool {
	return CanManageTable(p, meta, admin) || CanRotateInvites(p, meta, admin)
}

// SeatRemoved implements auth.SeatChecker: true when the game is live in
// this lobby and has no seat with that player id. A game the lobby does
// not hold says false: nothing is known, so nothing is refused.
func (l *Lobby) SeatRemoved(gameID, playerID uuid.UUID) bool {
	l.mu.Lock()
	entry, ok := l.games[gameID]
	l.mu.Unlock()
	if !ok {
		return false
	}
	return entry.room.Game.PlayerByID(playerID) == nil
}

var _ auth.SeatChecker = (*Lobby)(nil)

// RemoveSeat takes a non-host, non-bot seat off an unstarted table and
// returns the table as it now stands. The caller is authorised at the
// HTTP layer (CanRemoveSeat). Socket eviction is the HTTP layer's too:
// the lobby has no hub reference.
func (l *Lobby) RemoveSeat(id, playerID uuid.UUID) (GameMeta, error) {
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if entry.room.Game.CurrentState() != game.StateLobby {
		return GameMeta{}, ErrGameStarted
	}
	idx := -1
	for i, seat := range entry.meta.Players {
		if seat.PlayerID == playerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return GameMeta{}, ErrPlayerNotInGame
	}
	if entry.meta.Players[idx].IsBot {
		return GameMeta{}, ErrSeatIsBotRoute
	}
	// Read the host from the room, not the cached flag: hosting moves in
	// the engine and the meta mirrors it lazily.
	if entry.room.HostPlayerID() == playerID {
		return GameMeta{}, ErrCannotRemoveHost
	}
	var err error
	if broadcast, err = l.applyLocked(id, entry, func() error {
		return entry.room.Game.RemovePlayer(playerID)
	}); err != nil {
		if errors.Is(err, game.ErrGameNotInLobby) {
			return GameMeta{}, ErrGameStarted
		}
		return GameMeta{}, err
	}
	entry.meta.Players = append(entry.meta.Players[:idx], entry.meta.Players[idx+1:]...)
	for i := range entry.meta.Players {
		entry.meta.Players[i].Seat = i
	}
	l.syncHostLocked(entry)
	l.persistSeatsLocked(entry)
	return copyMeta(entry.meta), nil
}

// removeSeat handles DELETE /games/{id}/seats/{player}.
//
// The route shape: the bot routes are POST /games/{id}/seats/bot and
// DELETE /games/{id}/seats/bot/{player}, so the resource is "a seat at
// this table" and a seat is addressed by its player id. DELETE on that
// is the same verb on the same noun without the bot qualifier. The two
// DELETE patterns differ in segment count, so the mux never confuses
// them, and a bot's player id sent here is answered 422 with a pointer
// to the bot route rather than handled under this stricter gate.
//
//	401  anonymous
//	403  not the host, the creator or an admin; or player is not a seat here
//	404  no such table
//	409  the table has started
//	422  the seat is the host's, or a bot's
func removeSeat(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	playerID, err := uuid.Parse(r.PathValue("player"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid player id")
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	admin := c.isAdmin(p)
	if !CanRemoveSeat(p, meta, admin) {
		return ErrNotSeatManager
	}
	meta, err = c.Lobby.RemoveSeat(id, playerID)
	if err != nil {
		return err
	}
	// After the seat is gone, so a reconnect the closed client tries
	// straight away is already refused.
	if c.SeatEvictor != nil {
		c.SeatEvictor.EvictSeat(id, playerID)
	}
	return writeJSON(w, http.StatusOK, redactMetaFor(p, admin, id, meta))
}
