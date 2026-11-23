package game

import (
	"slices"

	"github.com/google/uuid"
)

// prev_life_lost.go: "lost life last turn" (Feast on the Fallen). The
// players who lost life during a turn are stamped into the next turn's
// tally as it begins, so the fact clones, snapshots and undoes with the
// rest of the tally (PrevNoncombatVictims' shape).

// lifeLosersThisTurnLocked lists, in seat order, the players the tally
// records as having lost life. Caller must hold g.mu.
func (g *Game) lifeLosersThisTurnLocked(t TurnTally) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && t.Players[p.ID].LifeLost > 0 && !slices.Contains(out, p.ID) {
			out = append(out, p.ID)
		}
	}
	return out
}

// LostLifeLastTurnForEffect reports whether `player` lost life during the
// previous turn, whoever's it was. False on the first turn of the game.
// Caller must hold g.mu.
func (g *Game) LostLifeLastTurnForEffect(player uuid.UUID) bool {
	return slices.Contains(g.TurnTally.PrevLifeLosers, player)
}
