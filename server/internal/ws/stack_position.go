package ws

// stack_position.go: where each seat has put the stack panel (#3429, ADR
// 0176).
//
// Like the voice room, the rematch offer and the playmats, the placement is
// not part of the game: the lobby sets it here and every captured view is
// stamped with it. Nothing about it enters the engine, its snapshot or its
// undo stack, and it lasts as long as the room does. The stamp is the FULL
// map, every seat's spot; protocol.FilterViewFor reduces it per viewer.

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// SetStackPosition records where seat playerID dragged the stack panel. A
// nil pos clears it, which puts the panel back at its default. Like
// SetVoice it takes only its own mutex, so the lobby may call it from
// inside an ApplyExternal fn.
func (r *Room) SetStackPosition(playerID uuid.UUID, pos *protocol.StackPositionView) {
	if playerID == uuid.Nil {
		return
	}
	r.stackPosMu.Lock()
	defer r.stackPosMu.Unlock()
	if pos == nil {
		delete(r.stackPos, playerID)
		return
	}
	if r.stackPos == nil {
		r.stackPos = map[uuid.UUID]protocol.StackPositionView{}
	}
	r.stackPos[playerID] = *pos
}

// StackPosition returns seat playerID's saved spot, nil when it has none.
func (r *Room) StackPosition(playerID uuid.UUID) *protocol.StackPositionView {
	r.stackPosMu.Lock()
	defer r.stackPosMu.Unlock()
	p, ok := r.stackPos[playerID]
	if !ok {
		return nil
	}
	return &p
}

// stampStackPositionsLocked writes every seat's spot onto a freshly
// captured view. Caller holds r.mu. An ended game carries none: the panel
// is gone with the stack.
func (r *Room) stampStackPositionsLocked(view *protocol.GameView) {
	r.stackPosMu.Lock()
	defer r.stackPosMu.Unlock()
	view.StackPositions = nil
	if len(r.stackPos) == 0 || view.State == "ended" {
		return
	}
	view.StackPositions = make(map[string]protocol.StackPositionView, len(r.stackPos))
	for id, p := range r.stackPos {
		view.StackPositions[id.String()] = p
	}
}
