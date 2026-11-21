package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b141_helpers.go — the "at the beginning of your upkeep, if you have
// N or more life, you win the game" family (Felidar Sovereign's 40, Test of
// Endurance's 50; #5948), moved out of Felidar Sovereign's card file so a card
// in another shard can call it.

// LifeAtLeast reports whether `player` is still in the game with at least `n`
// life: the intervening-if of that family.
func LifeAtLeast(g *game.Game, player uuid.UUID, n int) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && !p.Eliminated && p.Life >= n
}

// WinIfLifeAtLeast is that family's effect: the intervening-if checked again on
// resolution (CR 603.4), then WinTheGame. Captures only the threshold, so it
// survives Clone and undo.
func WinIfLifeAtLeast(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !LifeAtLeast(g, item.Controller, n) {
			return nil
		}
		return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
	}
}
