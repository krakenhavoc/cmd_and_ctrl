package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quirion Dryad — Creature — Dryad {1}{G}, 1/1 (Core Set 2021, #5511):
//
//	"Whenever you cast a spell that's white, blue, black, or red, put a +1/+1
//	 counter on this creature."
//
// One trigger per spell cast, read from the spell's colours as it is cast: a
// white-and-blue spell is one trigger, a green or colourless one none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f42b6510-7322-4733-85ea-7ae6bd6ca8fc",
		Name:         "Quirion Dryad",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(effects.Or(effects.OfColor("W"), effects.OfColor("U"), effects.OfColor("B"), effects.OfColor("R"))),
				"Quirion Dryad — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
