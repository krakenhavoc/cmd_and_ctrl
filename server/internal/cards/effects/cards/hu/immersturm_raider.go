package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Immersturm Raider — Creature — Demon Berserker {1}{R}, 2/1 (Kaldheim,
// #5505):
//
//	"When this creature enters, you may discard a card. If you do, draw
//	 a card."
//
// Fissure Wizard's enters trigger: the discard is up to one card, and
// the draw follows only a card actually discarded.
//
// No simplification.
func init() {
	const label = "Immersturm Raider — you may discard a card. If you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "e3179dd6-5558-4385-8883-f78f9d1b1c4f",
		Name:         "Immersturm Raider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}
