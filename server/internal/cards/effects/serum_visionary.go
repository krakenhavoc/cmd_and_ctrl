package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serum Visionary — Creature — Vedalken Wizard {2}{U}, 2/2 (Modern
// Horizons 3, #5467):
//
//	"When this creature enters, draw a card, then scry 2."
//
// Cavern Stomper's enters scry with a draw in front of it: one trigger
// whose effect is Do(DrawCards, Scry), so the scry looks at the library
// after the draw (CR 608.2c) and the footprinted primitives run in
// printed order. The draw and the scry rider are declared on the card's
// own Purpose, as Archive Dragon does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "922d9d79-8038-4185-9dfb-2d2d52fd8460",
		Name:         "Serum Visionary",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Serum Visionary — draw a card, then scry 2", Do(DrawCards{N: 1}, Scry{N: 2})),
		},
	})
}
