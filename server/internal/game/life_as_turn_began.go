package game

import "github.com/google/uuid"

// life_as_turn_began.go — each player's life total as the turn began
// (#6164): Sengir, the Dark Baron's "Whenever another player loses the
// game, you gain life equal to that player's life total as the turn
// began." The life a player had when the turn began is not on the board
// once the turn has moved it, so it is recorded as the turn begins, for
// every player in the game, on Player (carried by the snapshot).

// stampLifeAsTurnBeganLocked records every player's current life as the
// life they had as the turn began. Run from startTurnLocked, once per turn
// that really begins. Caller must hold g.mu.
func (g *Game) stampLifeAsTurnBeganLocked() {
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		p.LifeAsTurnBegan = p.Life
		p.LifeAsTurnBeganSeq = g.Turn.Seq
	}
}

// LifeAsTurnBeganForEffect is `player`'s life total as the current turn
// began. A player who has left the game keeps the record, so a "loses the
// game" trigger reads it after the loss. When no record was taken this
// turn (the game's first turn before any turn began, or a game restored
// from a file written before the record existed) it falls back to the
// player's current life. ok is false for an unknown player.
//
// Caller must hold either lock.
func (g *Game) LifeAsTurnBeganForEffect(player uuid.UUID) (int, bool) {
	for _, p := range g.Seats {
		if p == nil || p.ID != player {
			continue
		}
		if p.LifeAsTurnBeganSeq == g.Turn.Seq && p.LifeAsTurnBeganSeq != 0 {
			return p.LifeAsTurnBegan, true
		}
		return p.Life, true
	}
	return 0, false
}
