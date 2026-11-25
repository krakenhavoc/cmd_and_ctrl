package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thousand Winds — Creature — Elemental {4}{U}{U}, 5/6 (Commander 2019):
//
//	"Flying
//	 Morph {5}{U}{U} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its morph cost.)
//	 When this creature is turned face up, return all other tapped creatures
//	 to their owners' hands."
//
// Morph is the shared constructor, the trigger Belltoll Dragon's
// turned-face-up shape. The bounce is Waterspout Elemental's mass bounce over
// the tapped creatures, read as the trigger resolves (CR 608.2c), spares this
// creature, and is symmetric: yours go home too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "81b50d34-d778-4606-80de-0e0780610d1f",
		Name:             "Thousand Winds",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{5}{U}{U}")},
		Purpose:          game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Thousand Winds — return all other tapped creatures to their owners' hands",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BounceAllMatching{Match: effects.And(effects.HobB3TappedCreature(), effects.NotSelf(ctx.Source()))}.Apply(ctx)
				}),
		},
	})
}
