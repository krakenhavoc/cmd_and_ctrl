package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bellowing Crier — Creature — Frog Advisor {1}{U}, 2/1 (Bloomburrow,
// #4467):
//
//	"When this creature enters, draw a card, then discard a card."
//
// Cloistered Telepath's loot (b35LootOne): the draw first, so the card
// drawn is a legal discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2f3be8a-5ea1-47c2-8104-747f4589f5e6",
		Name:         "Bellowing Crier",
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bellowing Crier — draw a card, then discard a card", b35LootOne),
		},
	})
}
