package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloistered Telepath — Creature — Talosian Advisor {1}{U}, 2/1 (Star Trek):
//
//	Flash
//	When this creature enters, draw a card, then discard a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8656291d-8de7-4341-b1ee-0fcd65d629a4",
		Name:            "Cloistered Telepath",
		Purpose:         game.Purpose{Draws: 1, Discards: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cloistered Telepath — draw a card, then discard a card", effects.B35LootOne),
		},
	})
}
