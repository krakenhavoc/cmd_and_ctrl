package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloudpiercer — Creature — Dinosaur {4}{R}, 5/4 (Ikoria: Lair of Behemoths,
// #5512):
//
//	"Mutate {3}{R} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Reach
//	 Whenever this creature mutates, you may discard a card. If you do, draw a
//	 card."
//
// Mutate is the engine's (ADR 0223) and the loot is Rubble Rouser's shared
// discard-then-draw prompt: the draw follows only a discard that happened.
//
// No simplification.
func init() {
	const label = "Cloudpiercer — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:         "ed7885bf-c557-4328-b1a4-54780b2315f4",
		Name:             "Cloudpiercer",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"reach"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates, label, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}
