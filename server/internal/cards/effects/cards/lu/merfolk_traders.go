package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Traders — Creature — Merfolk {1}{U}, 1/2 (Weatherlight, #5629):
//
//	"When this creature enters, draw a card, then discard a card."
//
// Bellowing Crier's enters loot: the draw, then a discard the
// controller chooses (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8af6646a-10ba-448b-bd5e-48feb70de5e5",
		Name:         "Merfolk Traders",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Merfolk Traders — draw a card, then discard a card", effects.B35LootOne),
		},
	})
}
