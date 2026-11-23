package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helpers roll-up 40 moved out of card shards into the core, because cards
// in other shards need them too (a shard cannot import another shard).

// R40APlayerAttacksYou is "whenever a player attacks you" (Archnemesis,
// Mirkwood Trapper).
func R40APlayerAttacksYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventAttack || ev.Actor == source.Controller {
		return false
	}
	return AttackedPlayer(g, ev) == source.Controller
}

// R40LifeAtLeast reports whether `player` is still in the game with at
// least `n` life — the intervening-if of the "at the beginning of your
// upkeep, if you have N or more life, you win the game" family (Felidar
// Sovereign's 40, Test of Endurance's 50).
func R40LifeAtLeast(g *game.Game, player uuid.UUID, n int) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && !p.Eliminated && p.Life >= n
}

// R40WinIfLifeAtLeast is that family's effect: the intervening-if checked
// again on resolution (CR 603.4), then WinTheGame. Captures only the
// threshold, so it survives Clone and undo.
func R40WinIfLifeAtLeast(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !R40LifeAtLeast(g, item.Controller, n) {
			return nil
		}
		return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
	}
}
