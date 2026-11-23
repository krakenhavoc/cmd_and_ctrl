package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "099a5835-da6c-4e03-ad3e-aeb448897fed",
		Name:            "Owl Familiar",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Draws: 1, Discards: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Owl Familiar — draw a card, then discard a card", effects.B35LootOne),
		},
	})
}
