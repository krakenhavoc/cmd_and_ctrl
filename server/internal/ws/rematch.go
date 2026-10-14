package ws

// rematch.go: the host's rematch offer on the ended table's game view
// (#3285, ADR 0164).
//
// Like the voice room, the host and the playmats, the offer is not part
// of the game: the lobby sets it here and every captured view is stamped
// with it. Nothing about it enters the engine, its snapshot or its undo
// stack. The stamp is the FULL state, every seat's answer included;
// protocol.FilterViewFor reduces it per viewer.

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// SetRematch replaces the room's rematch offer. nil clears it. The value
// is copied, so the caller may keep and change its own. Like SetVoice it
// takes only its own mutex, so the lobby may call it from inside an
// ApplyExternal fn.
func (r *Room) SetRematch(v *protocol.RematchView) {
	r.rematchMu.Lock()
	defer r.rematchMu.Unlock()
	r.rematch = copyRematch(v)
}

// Rematch returns a copy of the room's rematch offer, nil when unset.
func (r *Room) Rematch() *protocol.RematchView {
	r.rematchMu.Lock()
	defer r.rematchMu.Unlock()
	return copyRematch(r.rematch)
}

// stampRematchLocked writes the offer onto a freshly captured view.
// Caller holds r.mu.
func (r *Room) stampRematchLocked(view *protocol.GameView) {
	r.rematchMu.Lock()
	defer r.rematchMu.Unlock()
	view.Rematch = copyRematch(r.rematch)
}

func copyRematch(v *protocol.RematchView) *protocol.RematchView {
	if v == nil {
		return nil
	}
	out := *v
	out.Accepted = append([]string(nil), v.Accepted...)
	out.Left = append([]string(nil), v.Left...)
	return &out
}
