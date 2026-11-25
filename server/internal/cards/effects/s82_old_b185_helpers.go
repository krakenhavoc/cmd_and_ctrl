package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b185_helpers.go — Guilds of Ravnica, S82 older sets batch 185 (#5525).

// S82OldB185YouveSurveilledThisTurn is "you've surveilled this
// turn": an EventSurveil with `player` as the actor. A scry is not a surveil.
// Neither is emitted for one that looked at nothing, because none happened.
// Caller holds g.mu.
func S82OldB185YouveSurveilledThisTurn(g *game.Game, player uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventSurveil && ev.Actor == player {
			return true
		}
	}
	return false
}
