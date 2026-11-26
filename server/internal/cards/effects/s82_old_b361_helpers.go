package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b361_helpers.go — S82 OLD batch 361 (Conspiracy, #5557).

// S82OldB361FirstUpkeepOfTheGame is the When of "at the beginning of the first
// upkeep": the upkeep step that is beginning is its player's first
// (Player.UpkeepsBegun counts it as the step begins) and no other player has
// begun an upkeep yet, so it is the game's first upkeep. The trigger is the
// conspiracy's, wherever the active player sits.
func S82OldB361FirstUpkeepOfTheGame(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBeginUpkeep {
		return false
	}
	total := 0
	for _, p := range g.Seats {
		total += p.UpkeepsBegun
	}
	p := g.PlayerByIDForEffect(ev.Actor)
	return p != nil && p.UpkeepsBegun == 1 && total == 1
}
