package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Traders — Creature — Merfolk {1}{U}, 1/2 (Weatherlight, #5629):
//
//	"When this creature enters, draw a card, then discard a card."
//
// Bellowing Crier's enters loot: the draw, then a discard the
// controller chooses (CR 701.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8af6646a-10ba-448b-bd5e-48feb70de5e5",
		Name:         "Merfolk Traders",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Merfolk Traders — draw a card, then discard a card", b35LootOne),
		},
	})
}
