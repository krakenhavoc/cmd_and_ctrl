package lobby

// playmat.go: the lobby's half of playmats (ADR 0128): putting a
// seated person's playmat URL on their seat, and moving it when they
// change it.
//
// The image, its storage and its routes live in internal/playmat and
// playmat_http.go. What lives here is only the seat binding. The room
// keeps seat -> URL (ws.Room.SetPlaymat) and stamps it on every
// captured view, so the engine never sees a playmat.
//
// A seat gets one when a SIGNED-IN person holds it: a guest has no
// account to own a mat, a bot has none, and an agent seat is a guest.

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// PlaymatSource is the one question the lobby asks of the playmat
// service: the same-origin URL of an account's playmat, or "" for none.
// *playmat.Service implements it.
type PlaymatSource interface {
	URL(user uuid.UUID) string
}

// SetPlaymats wires the playmat source in after construction. Call it
// before RestoreFromDisk so a resumed table's seats get their mats.
func (l *Lobby) SetPlaymats(p PlaymatSource) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.playmats = p
}

// bindPlaymat puts userID's playmat on the seat playerID of room. A
// zero userID (a guest) or no source leaves the seat without one. Safe
// inside an ApplyExternal fn: it takes only the room's playmat mutex
// and the source's own lock, never l.mu or the room's.
func (l *Lobby) bindPlaymat(room *ws.Room, playerID, userID uuid.UUID) {
	if l.playmats == nil || userID == uuid.Nil || room == nil {
		return
	}
	room.SetPlaymat(playerID, l.playmats.URL(userID))
	l.bindPlaymatWash(room, playerID, userID)
}

// PlaymatChanged moves a person's new playmat onto every table they
// hold a seat at, and pushes the change to the connected clients. url
// is "" when it was removed. Called by the playmat routes after the
// image is stored or deleted, and by an admin removing one.
//
// A change is a lobby step, not a game action: it records no undo
// entry and no log line, and it reaches the table on the same commit
// path a join does. Ended and archived tables are skipped; nobody is
// looking at them live, and a table that is opened again is stamped
// from the account afresh.
func (l *Lobby) PlaymatChanged(userID uuid.UUID, url string) {
	if userID == uuid.Nil {
		return
	}
	want := userID.String()
	var broadcasts []func()
	l.mu.Lock()
	for id, entry := range l.games {
		if entry.meta.Archived() || entry.room == nil || entry.room.Game.CurrentState() == game.StateEnded {
			continue
		}
		for _, s := range entry.meta.Players {
			if s.UserID != want || s.IsBot || s.IsAgent {
				continue
			}
			pid := s.PlayerID
			b, err := l.applyLocked(id, entry, func() error {
				entry.room.SetPlaymat(pid, url)
				l.bindPlaymatWash(entry.room, pid, userID)
				return nil
			})
			if err != nil {
				continue
			}
			broadcasts = append(broadcasts, b)
		}
	}
	l.mu.Unlock()
	for _, b := range broadcasts {
		b()
	}
}
