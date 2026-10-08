package ws

// playmat.go: the playmat each seat shows behind its battlefield (ADR
// 0128).
//
// A playmat belongs to an ACCOUNT, not to the engine's Player: the
// player's owner can change it in Settings at any moment, and the
// table must see the change on the next snapshot without a game
// action. So it is kept here, like the host, and stamped on every
// captured view. Nothing about it enters the engine, its snapshot
// schema or its undo stack.
//
// The lobby sets it (SetPlaymat) when a signed-in person sits down or
// reclaims a seat, when a restored room is paired with its metadata,
// and when the person changes it. Bot and agent seats are never given
// one, and the stamp skips them even if one were set.

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// SetPlaymat records the same-origin URL of the playmat seat playerID
// shows. An empty url clears it. Like SetHost it takes only its own
// mutex, so the lobby may call it from inside an ApplyExternal fn,
// which is how a join makes its own capture carry the mat.
func (r *Room) SetPlaymat(playerID uuid.UUID, url string) {
	if playerID == uuid.Nil {
		return
	}
	r.playmatMu.Lock()
	defer r.playmatMu.Unlock()
	if url == "" {
		delete(r.playmats, playerID)
		return
	}
	if r.playmats == nil {
		r.playmats = map[uuid.UUID]string{}
	}
	r.playmats[playerID] = url
}

// SetPlaymatWash records the owner-set wash of seat playerID's playmat
// (ADR 0128 amendment). 0 clears it, and the stamp then sends none, so
// the client uses its default. Same locking as SetPlaymat.
func (r *Room) SetPlaymatWash(playerID uuid.UUID, wash int) {
	if playerID == uuid.Nil {
		return
	}
	r.playmatMu.Lock()
	defer r.playmatMu.Unlock()
	if wash <= 0 {
		delete(r.playmatWashes, playerID)
		return
	}
	if r.playmatWashes == nil {
		r.playmatWashes = map[uuid.UUID]int{}
	}
	r.playmatWashes[playerID] = wash
}

// stampPlaymatsLocked writes each human seat's playmat onto a freshly
// captured view. Caller holds r.mu. The same URL goes to every viewer:
// a playmat is public to the table, like the seat's name.
func (r *Room) stampPlaymatsLocked(view *protocol.GameView) {
	r.playmatMu.Lock()
	defer r.playmatMu.Unlock()
	for i := range view.Seats {
		s := &view.Seats[i]
		s.PlaymatURL = ""
		s.PlaymatWash = 0
		if len(r.playmats) == 0 || s.IsBot || s.IsAgent {
			continue
		}
		id, err := uuid.Parse(s.ID)
		if err != nil {
			continue
		}
		s.PlaymatURL = r.playmats[id]
		if s.PlaymatURL != "" {
			s.PlaymatWash = r.playmatWashes[id]
		}
	}
}
