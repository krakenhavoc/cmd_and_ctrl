package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vaultbreaker — Creature — Orc Rogue {3}{R}, 4/2 (Fate Reforged,
// #5552):
//
//	"Whenever this creature attacks, you may discard a card. If you do,
//	 draw a card.
//	 Dash {2}{R} (You may cast this spell for its dash cost. If you do,
//	 it gains haste, and it's returned from the battlefield to its
//	 owner's hand at the beginning of the next end step.)"
//
// The attack trigger is Burning-Tree Vandal's rummage: the discard is
// optional, and only a card actually discarded draws one. Dash is the shared
// Dash constructor (freerunning.go): an alternative cost whose entry clause
// gives haste and returns the creature to its owner's hand at the beginning
// of the next end step (CR 702.109a, game/dash.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2577860a-663a-4b56-880c-05576522e3b7",
		Name:             "Vaultbreaker",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(vaultbreakerLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, vaultbreakerLabel, effects.DrawAsManyAsDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}

const vaultbreakerLabel = "Vaultbreaker — you may discard a card; if you do, draw a card"
