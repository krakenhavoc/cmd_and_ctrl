package lobby

// stack_position.go is the lobby's half of the movable stack panel (#3429,
// ADR 0176): a seat drags the stack somewhere else on its board, and the
// table remembers where for the rest of the game.
//
// The placement is a seat's own and only ever reaches that seat's view. The
// room keeps it (ws.Room.SetStackPosition) and stamps it on every captured
// view, so the engine never sees it: it is not in the snapshot, `undo` does
// not touch it, and it goes with the room. It is not written to disk, so a
// server restart puts every panel back at its default.

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

var (
	// ErrBadStackPosition refuses a placement that is not two fractions
	// of the board, each from 0 to 1.
	ErrBadStackPosition = errors.New("lobby: the stack position must be two fractions between 0 and 1")
	// ErrStackPositionNotSeated refuses a placement for a seat that is not
	// a human at this table.
	ErrStackPositionNotSeated = errors.New("lobby: only a player seated at this table can place its stack")
	// ErrStackPositionEnded refuses a placement on a table that is over.
	ErrStackPositionEnded = errors.New("lobby: the game is over")
)

// SetStackPosition records where seat playerID put its stack panel and
// pushes it to that seat's clients. A nil pos is the reset: the saved spot
// is dropped and the panel goes back to its default. Resetting a seat that
// has none is not an error.
func (l *Lobby) SetStackPosition(id, playerID uuid.UUID, pos *protocol.StackPositionView) error {
	if pos != nil && !protocol.ValidStackPosition(*pos) {
		return ErrBadStackPosition
	}
	l.mu.Lock()
	entry, ok := l.games[id]
	if !ok {
		l.mu.Unlock()
		return ErrGameNotFound
	}
	seated := false
	for _, s := range entry.meta.Players {
		if s.PlayerID == playerID && !s.IsBot && !s.IsAgent {
			seated = true
		}
	}
	if !seated {
		l.mu.Unlock()
		return ErrStackPositionNotSeated
	}
	if entry.room.Game.CurrentState() == game.StateEnded {
		l.mu.Unlock()
		return ErrStackPositionEnded
	}
	broadcast, err := l.applyLocked(id, entry, func() error {
		entry.room.SetStackPosition(playerID, pos)
		return nil
	})
	l.mu.Unlock()
	if err != nil {
		return fmt.Errorf("lobby: saving the stack position: %w", err)
	}
	broadcast()
	return nil
}
