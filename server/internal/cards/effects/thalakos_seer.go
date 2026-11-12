package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thalakos Seer — Creature — Thalakos Wizard {U}{U}, 1/1
// (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 When this creature leaves the battlefield, draw a card."
//
// Shadow is a printed keyword. The draw is Circuit Mender's
// leaves-the-battlefield trigger (CR 603.6c): dying, exile, a bounce or
// a sacrifice all count, and the Seer's controller as it left draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1baec770-0785-4cbc-be18-83572fd8e9c1",
		Name:            "Thalakos Seer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, Self, "Thalakos Seer — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
