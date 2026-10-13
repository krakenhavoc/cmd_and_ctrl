package lobby

// seat_remove.go: the host removes someone else's seat from an
// unstarted table (#2785).
//
// Until this existed only a bot seat could be taken off a table
// (DELETE /games/{id}/seats/bot/{player}). A human or agent guest who
// sat down under the wrong name could not be moved, and the only way
// out was a new table and a new invite.
//
// SCOPE: unstarted tables for the host, the creator and an admin, and
// STARTED tables for an admin alone (#3093, the owner's call of
// 2026-10-12). The host and the creator get a 409 on a started table,
// as before, and the engine's RemovePlayer refuses it again underneath.
//
// WHO: the table host, the game's creator, or an admin. That is the
// set that manages the table (CanManageTable) and the invites
// (CanRotateInvites) put together. It is narrower than the bot route,
// which any seated player may call: removing a person is not the same
// kind of act as adding a bot, and no one gains it who did not have it.
//
// WHAT A REMOVAL FROM A STARTED GAME DOES (RemoveStartedSeat):
//
//   - the seat concedes, through game.RemoveSeatFromGame, which is
//     Concede: CR 800.4a runs exactly as it does for a player who gives
//     up (their objects leave, control effects end, their stack items
//     cease to exist, the turn and priority move on, one player left
//     ends the game). The log records a concession;
//   - the seat is NOT taken out of the seat list. The game is under way,
//     the others have been looking at that player, and the engine keeps
//     the eliminated seat in turn order; it reads as a player who left;
//   - the engine marks the seat Removed, which is what the session check
//     reads (below). It rides the snapshot, so a restart keeps it;
//   - the undo ring is emptied. Everything on it is a clone from before
//     the seat left, and undoing one would put the player back at a
//     table that has already refused them;
//   - a removed player does not keep spectating: their session is
//     refused like an unstarted removal's. A person holding the
//     spectator invite can still open the table as any spectator can.
//
// WHAT A REMOVAL FROM AN UNSTARTED TABLE DOES:
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
// invite link can sit down again, as a new seat, UNLESS the remover also
// blocked them (?block=true, blocks.go, #2984): then the same account, or
// a guest of the same name, is refused with a clear message. Rotating the
// invite is the way to stop everyone else, and the Remove control says so.

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

	// ErrGameNotInProgress refuses removing a seat from a game that is
	// not under way: an ended game has nothing left to leave (a table
	// still in the lobby has RemoveSeat).
	ErrGameNotInProgress = errors.New("lobby: the game is not in progress; there is no seat to concede out of it")
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
// this lobby and either has no seat with that player id (an unstarted
// table's removal drops the seat) or has one an admin removed from a
// started game (game.Player.Removed). A game the lobby does not hold
// says false: nothing is known, so nothing is refused.
func (l *Lobby) SeatRemoved(gameID, playerID uuid.UUID) bool {
	l.mu.Lock()
	entry, ok := l.games[gameID]
	l.mu.Unlock()
	if !ok {
		return false
	}
	return entry.room.Game.SeatRemoved(playerID)
}

var _ auth.SeatChecker = (*Lobby)(nil)

// seatRemovedAt reports whether the seat at index seat of entry's list
// was removed from a started game, so the rejoin link of "my games" is
// not offered to someone whose session would be refused. Caller holds
// l.mu.
func seatRemovedAt(entry *gameEntry, seat int) bool {
	for _, s := range entry.meta.Players {
		if s.Seat == seat {
			return entry.room.Game.SeatRemoved(s.PlayerID)
		}
	}
	return false
}

// RemoveSeat takes a non-host, non-bot seat off an unstarted table and
// returns the table as it now stands. The caller is authorised at the
// HTTP layer (CanRemoveSeat). Socket eviction is the HTTP layer's too:
// the lobby has no hub reference.
func (l *Lobby) RemoveSeat(id, playerID uuid.UUID) (GameMeta, error) {
	return l.RemoveSeatBlocking(id, playerID, false)
}

// RemoveSeatBlocking is RemoveSeat that also, when block is true, blocks
// the removed seat from taking a seat at this table again (blocks.go).
func (l *Lobby) RemoveSeatBlocking(id, playerID uuid.UUID, block bool) (GameMeta, error) {
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
	if block {
		l.addBlocksLocked(entry, blockKeysOf(entry.meta.Players[idx]))
	}
	entry.meta.Players = append(entry.meta.Players[:idx], entry.meta.Players[idx+1:]...)
	for i := range entry.meta.Players {
		entry.meta.Players[i].Seat = i
	}
	l.syncHostLocked(entry)
	l.persistSeatsLocked(entry)
	return copyMeta(entry.meta), nil
}

// RemoveStartedSeat takes a non-host, non-bot seat out of a game that is
// under way by conceding it (#3093) and returns the table as it now
// stands. The caller has checked at the HTTP layer that the requester
// is an admin: this function has no notion of who is asking. The seat
// stays in the list; see the file comment for what the removal does.
func (l *Lobby) RemoveStartedSeat(id, playerID uuid.UUID) (GameMeta, error) {
	return l.RemoveStartedSeatBlocking(id, playerID, false)
}

// RemoveStartedSeatBlocking is RemoveStartedSeat that also blocks the
// seat's identities from this table, when block is true. A started table
// takes no new seats, so the block matters if it is ever restarted from
// a restore point or the person is later asked to spectate, and it keeps
// the record the admin made.
func (l *Lobby) RemoveStartedSeatBlocking(id, playerID uuid.UUID, block bool) (GameMeta, error) {
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
	if entry.room.Game.CurrentState() != game.StateActive {
		return GameMeta{}, ErrGameNotInProgress
	}
	seat, ok := findSeat(entry.meta.Players, playerID)
	if !ok {
		return GameMeta{}, ErrPlayerNotInGame
	}
	if seat.IsBot {
		return GameMeta{}, ErrSeatIsBotRoute
	}
	if entry.room.HostPlayerID() == playerID {
		return GameMeta{}, ErrCannotRemoveHost
	}
	view, seq, err := entry.room.ApplyExternalClearingUndo(func() error {
		return entry.room.Game.RemoveSeatFromGame(playerID)
	})
	if err != nil {
		if errors.Is(err, game.ErrGameNotActive) {
			return GameMeta{}, ErrGameNotInProgress
		}
		return GameMeta{}, err
	}
	if block {
		l.addBlocksLocked(entry, blockKeysOf(seat))
	}
	if l.broadcast != nil {
		b := l.broadcast
		broadcast = func() { b.BroadcastState(id, seq, view) }
	}
	// The concession can end the game; record it the way every other
	// path that ends one does.
	l.syncStateLocked(entry)
	l.syncHostLocked(entry)
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
// A table that has started takes the admin only (#3093): the host and
// the creator are answered 409, as they were before the admin could.
//
//	401  anonymous
//	403  not the host, the creator or an admin; or player is not a seat here
//	404  no such table
//	409  the table has started and the caller is not an admin; or the game
//	     has ended
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
	// ?block=true also blocks the removed seat from this table (#2984).
	block := r.URL.Query().Get("block") == "true"
	if meta.State == string(game.StateLobby) {
		meta, err = c.Lobby.RemoveSeatBlocking(id, playerID, block)
	} else if !admin {
		return ErrGameStarted
	} else {
		meta, err = c.Lobby.RemoveStartedSeatBlocking(id, playerID, block)
	}
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
