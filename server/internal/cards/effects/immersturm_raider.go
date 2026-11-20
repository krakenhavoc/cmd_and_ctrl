package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e3179dd6-5558-4385-8883-f78f9d1b1c4f",
		Name:         "Immersturm Raider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
