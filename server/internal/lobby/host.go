package lobby

// host.go: the table host (ADR 0075 §2.1).
//
// Who hosts:
//
//   - POST /games may name a host by Discord ID (host_discord_id; the
//     Discord bot's /c2-invite passes whoever ran it). It is held
//     pending on the meta until that identity claims a seat, and then
//     bound to that seat.
//   - Until then — and at every table that names nobody — the first
//     HUMAN seat to join hosts. A named host who arrives later takes
//     over from that interim host; the table is never left without
//     someone who can manage it just because the named host is late.
//   - Bot seats never host.
//   - POST /games/{id}/host transfers it (host or admin).
//   - A host who leaves the game (ADR 0060 — concede or loss) passes
//     hosting to the next human seat in turn order. That rule lives on
//     the room (ws/host.go) because it reads engine state; the lobby
//     mirrors the room's answer into GameMeta.HostPlayerID.

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ErrHostIneligible is returned by TransferHost when the target seat
// cannot host: a bot, or a player who has already left the game.
var ErrHostIneligible = errors.New("lobby: seat cannot host (bot or no longer in the game)")

// ErrNotTableManager is returned by the host-or-admin routes when the
// caller is neither.
var ErrNotTableManager = errors.New("lobby: only the table host or the admin may do that")

// CanManageTable is the one "host or admin" predicate (ADR 0075
// §2.1). True for the server admin, or for a player session bound to
// this game whose seat is the table's host. meta must be fresh from
// Lobby.Get (which resolves the effective host); a spectator, an
// unseated Discord sign-in, another seat, or the host of a DIFFERENT
// game is refused.
//
// The WebSocket side asks the same question through
// ws.Room.CanManageTable, which reads the same host.
func CanManageTable(p auth.Principal, meta GameMeta) bool {
	switch p.Role {
	case auth.RoleAdmin:
		return true
	case auth.RolePlayer:
		return p.PlayerID != uuid.Nil &&
			meta.HostPlayerID != uuid.Nil &&
			p.GameID == meta.ID &&
			p.PlayerID == meta.HostPlayerID
	default:
		return false
	}
}

// syncHostLocked mirrors the room's effective host onto the meta and
// stamps SeatInfo.IsHost. Writes the games row (host_player_id) when
// hosting has moved since the last write — a pass on elimination is
// observed here, not pushed. The host's twin of syncStateLocked.
// Caller holds l.mu.
func (l *Lobby) syncHostLocked(entry *gameEntry) {
	if entry == nil || entry.room == nil {
		return
	}
	h := entry.room.HostPlayerID()
	changed := h != entry.meta.HostPlayerID
	entry.meta.HostPlayerID = h
	for i := range entry.meta.Players {
		is := h != uuid.Nil && entry.meta.Players[i].PlayerID == h
		if entry.meta.Players[i].IsHost != is {
			changed = true
		}
		entry.meta.Players[i].IsHost = is
	}
	if changed {
		l.persistGameLocked(entry)
	}
}

// TransferHost makes target the table host. The caller is authorised
// at the HTTP layer (CanManageTable). target must be a human seat
// still in the game. An explicit transfer also drops any pending
// named host, so a late /c2-invite claimant does not take the table
// back from whoever it was handed to.
func (l *Lobby) TransferHost(id, target uuid.UUID) (GameMeta, error) {
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
	var seat *SeatInfo
	for i := range entry.meta.Players {
		if entry.meta.Players[i].PlayerID == target {
			seat = &entry.meta.Players[i]
			break
		}
	}
	if seat == nil {
		return GameMeta{}, ErrPlayerNotInGame
	}
	if seat.IsBot {
		return GameMeta{}, ErrHostIneligible
	}
	if !canHost(entry.room.Game, target) {
		return GameMeta{}, ErrHostIneligible
	}

	// Through the room so every connected client's is_host flag moves
	// with it; the capture stamps the new host.
	var err error
	if broadcast, err = l.applyLocked(id, entry, func() error {
		entry.room.SetHost(target)
		return nil
	}); err != nil {
		return GameMeta{}, err
	}
	entry.meta.HostDiscordID = ""
	l.syncHostLocked(entry)
	l.persistGameLocked(entry)
	return copyMeta(entry.meta), nil
}

// canHost reports whether the engine seat is a human still in the
// game. Read under the game's lock: Eliminated is written by the
// engine mid-game.
func canHost(g *game.Game, playerID uuid.UUID) bool {
	ok := false
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p != nil && p.ID == playerID {
				ok = !p.IsBot && !p.Eliminated
				return
			}
		}
	})
	return ok
}
