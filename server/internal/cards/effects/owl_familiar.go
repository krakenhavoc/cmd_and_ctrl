package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Owl Familiar — Creature — Bird {1}{U}, 1/1 (Portal, #5630):
//
//	"Flying
//	 When this creature enters, draw a card, then discard a card."
//
// Icewind Elemental's shape: flying (CR 702.9b) and an enters trigger
// that loots one (b35LootOne).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "099a5835-da6c-4e03-ad3e-aeb448897fed",
		Name:            "Owl Familiar",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Draws: 1, Discards: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Owl Familiar — draw a card, then discard a card", b35LootOne),
		},
	})
}
