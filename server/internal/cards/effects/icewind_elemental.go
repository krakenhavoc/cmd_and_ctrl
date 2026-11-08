package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Icewind Elemental — Creature — Elemental {4}{U}, 3/4 (Foundations,
// #4465):
//
//	"Flying
//	 When this creature enters, draw a card, then discard a card."
//
// Bellowing Crier's entry loot (b35LootOne) on a flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5001bf21-9dbb-4229-aeef-25805e71b832",
		Name:            "Icewind Elemental",
		Purpose:         game.Purpose{Draws: 1, Discards: 1},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Icewind Elemental — draw a card, then discard a card", b35LootOne),
		},
	})
}
