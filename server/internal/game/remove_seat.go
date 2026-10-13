package game

import "github.com/google/uuid"

// remove_seat.go: an admin takes a seat out of a game that has
// started (#3093, ADR 0060's 2026-10-12 amendment).
//
// This is a concession made on the seat's behalf and nothing more: it
// runs the same path as Concede, so CR 800.4a is whatever Concede
// already enforces — everything the seat owns leaves the game, effects
// that gave it control end, its spells and abilities on the stack
// cease to exist, its prompts are dropped, and the turn and priority
// move on if they were its. The game ends by the usual win check when
// one player is left. The log records a concession (EventConcede, then
// EventPlayerEliminated labelled "concede"): there is no separate
// "removed" line, because the table sees a player leave the game and
// the reason is the admin's, not the game's.
//
// What this adds is the record that the SEAT was taken away
// (Player.Removed), which the lobby reads to refuse the seat's session.
// A voluntary concession leaves it false, so a player who gives up
// still holds a credential that may spectate their own game.

// RemoveSeatFromGame removes playerID from an active game. A seat that
// has already left (conceded, or lost) is not conceded a second time,
// only marked, so the admin can still take the seat away from a player
// who has gone. Returns ErrGameNotActive outside an active game (a
// lobby has RemovePlayer, and an ended game has nothing left to leave)
// and ErrPlayerNotFound for an id that is not a seat.
//
// Not a player action: it has no wire type and no route into
// actions.Dispatch. It is called by the lobby, which has already
// checked that the caller is an admin.
func (g *Game) RemoveSeatFromGame(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if !p.Eliminated {
		g.concedeLocked(p)
	}
	p.Removed = true
	return nil
}

// SeatMarkedRemoved reports whether playerID is a seat an admin removed
// from a started game (Player.Removed). Unlike SeatRemoved it says false
// for a player the game does not have, so a caller labelling seats it
// already knows exist (the admin views) never reads a mismatch as a
// removal.
func (g *Game) SeatMarkedRemoved(playerID uuid.UUID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	p := g.playerByIDLocked(playerID)
	return p != nil && p.Removed
}

// SeatRemoved reports whether an admin removed playerID's seat from a
// started game (Player.Removed), or whether the game has no such seat.
// It reads under the game's lock, which a bare PlayerByID does not.
func (g *Game) SeatRemoved(playerID uuid.UUID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	p := g.playerByIDLocked(playerID)
	return p == nil || p.Removed
}
