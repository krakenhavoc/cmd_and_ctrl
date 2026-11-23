package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Axgard Artisan — Creature — Dwarf Artificer {1}{R}, 2/1 (March of the
// Machine):
//
//	"Whenever one or more +1/+1 counters are put on this creature for the
//	 first time each turn, create a Treasure token."
//
// One trigger per placement that raised the +1/+1 count and is the first
// this turn: Game.TimesPlusOneCountersPutThisTurn (per object, so a
// permanent that left and came back is new, CR 400.7) reads 1 as the
// trigger fires. Counters of another kind do not count, and a second
// placement the same turn triggers nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce8ec966-ecff-425e-944c-36dc8d4aef23",
		Name:         "Axgard Artisan",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Target == source.InstanceID && effects.MomB05PlusOneCountersFirstTimeThisTurn(ev, g)
			}, "Axgard Artisan — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
