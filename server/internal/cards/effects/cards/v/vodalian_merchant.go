package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vodalian Merchant — Creature — Merfolk {1}{U}, 1/2 (Invasion, S82 #5616):
//
//	"When this creature enters, draw a card, then discard a card."
//
// A loot on the enter trigger: the draw first, then the controller's own choice
// of discard (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d62c8f8-7a1c-4771-891d-05fdaa56bf47",
		Name:         "Vodalian Merchant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vodalian Merchant — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) }),
		},
	})
}
