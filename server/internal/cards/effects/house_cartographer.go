package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// House Cartographer — Creature — Human Scout Survivor {1}{G}, 2/2
// (Duskmourn: House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if this
//	 creature is tapped, reveal cards from the top of your library until
//	 you reveal a land card. Put that card into your hand and the rest on
//	 the bottom of your library in a random order."
//
// The Survival trigger (survival.go) with the shared reveal-until body
// (socB30RevealUntilToHandRestBottom) for a land: every card revealed is
// shown to the table (CR 701.20a), the land goes to the hand, and the rest
// go under in a random order. A library with no land reveals itself whole
// and every card goes to the bottom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76d93d50-4fc1-4ed4-a6dc-5fb193286b54",
		Name:         "House Cartographer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Survival("House Cartographer — reveal until a land card and put it into your hand",
				socB30RevealUntilToHandRestBottom("House Cartographer", func(c game.Card) bool { return c.IsLand() })),
		},
	})
}
