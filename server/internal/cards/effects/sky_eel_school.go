package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sky-Eel School — Creature — Fish {3}{U}{U}, 3/3 (Foundations Jumpstart):
//
//	"Flying
//	 When this creature enters, draw a card, then discard a card."
//
// The loot is the shared draw-then-discard body (b16DrawThenDiscard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c90ea9d5-82f7-4b92-8cb1-10d0a7378e8c",
		Name:            "Sky-Eel School",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sky-Eel School — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 1) }),
		},
	})
}
